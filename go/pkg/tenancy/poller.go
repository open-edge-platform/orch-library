// SPDX-FileCopyrightText: 2026 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package tenancy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// PollerConfig controls Poller behavior.
type PollerConfig struct {
	// PollInterval is the steady-state polling interval (default 5s).
	PollInterval time.Duration

	// PollLimit is the max number of events per poll request (default 100).
	PollLimit int

	// InitialBackoff is the starting backoff when the Tenant Manager is
	// unreachable (default 1s).
	InitialBackoff time.Duration

	// MaxBackoff caps the exponential backoff (default 30s).
	MaxBackoff time.Duration

	// OnError is called when a non-fatal error occurs (poll failure,
	// event processing error, status update failure). If nil, errors
	// are silently ignored. The Poller continues regardless.
	OnError func(err error, msg string)
}

// DefaultPollerConfig returns sensible defaults.
func DefaultPollerConfig() PollerConfig {
	return PollerConfig{
		PollInterval:   5 * time.Second,
		PollLimit:      100,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
	}
}

// Poller manages the full Watch lifecycle: replay on startup, then
// steady-state polling. It calls the Handler for each event and manages
// controller status updates automatically.
type Poller struct {
	tenantManagerURL string
	controllerName   string
	handler          Handler
	config           PollerConfig
	client           *http.Client
}

// NewPoller creates a Poller. controllerName must be the canonical ID
// from the registered-controller config (e.g., "app-orch-tenant-controller").
func NewPoller(tenantManagerURL, controllerName string, handler Handler, opts ...func(*PollerConfig)) *Poller {
	cfg := DefaultPollerConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Poller{
		tenantManagerURL: tenantManagerURL,
		controllerName:   controllerName,
		handler:          handler,
		config:           cfg,
		client:           &http.Client{Timeout: 30 * time.Second},
	}
}

// Run executes replay (Phase 1), then enters steady-state polling
// (Phase 2). Blocks until ctx is cancelled. On restart, replays from
// scratch. Handlers must be idempotent.
func (p *Poller) Run(ctx context.Context) error {
	// Phase 1: Replay with backoff retry.
	lastEventID, err := p.replayWithRetry(ctx)
	if err != nil {
		return err
	}

	// Phase 2: Steady-state polling.
	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			newLastID, err := p.poll(ctx, lastEventID)
			if err != nil {
				p.logError(err, "poll failed, will retry next interval")
				continue
			}
			lastEventID = newLastID
		}
	}
}

// replayWithRetry retries the replay request with exponential backoff
// until the Tenant Manager is available.
func (p *Poller) replayWithRetry(ctx context.Context) (int64, error) {
	backoff := p.config.InitialBackoff

	for {
		lastEventID, err := p.replay(ctx)
		if err == nil {
			return lastEventID, nil
		}

		p.logError(err, fmt.Sprintf("replay failed, retrying in %s", backoff))

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(backoff):
		}

		backoff *= 2
		if backoff > p.config.MaxBackoff {
			backoff = p.config.MaxBackoff
		}
	}
}

// replay fetches synthesized events and processes them.
func (p *Poller) replay(ctx context.Context) (int64, error) {
	url := fmt.Sprintf("%s/v1/events?controller=%s&replay=true",
		p.tenantManagerURL, p.controllerName)

	resp, err := p.doGet(ctx, url)
	if err != nil {
		return 0, err
	}

	for _, event := range resp.Events {
		if err := p.processEvent(ctx, event); err != nil {
			p.logError(err, fmt.Sprintf("replay event %s %s/%s failed",
				event.EventType, event.ResourceType, event.ResourceName))
		}
	}

	return resp.LastEventID, nil
}

// poll fetches incremental events after lastEventID and processes them.
func (p *Poller) poll(ctx context.Context, lastEventID int64) (int64, error) {
	url := fmt.Sprintf("%s/v1/events?controller=%s&after=%d&limit=%d",
		p.tenantManagerURL, p.controllerName, lastEventID, p.config.PollLimit)

	resp, err := p.doGet(ctx, url)
	if err != nil {
		return lastEventID, err
	}

	for _, event := range resp.Events {
		if err := p.processEvent(ctx, event); err != nil {
			p.logError(err, fmt.Sprintf("event %s %s/%s failed",
				event.EventType, event.ResourceType, event.ResourceName))
		}
	}

	return resp.LastEventID, nil
}

// processEvent handles a single event: set in_progress, call handler,
// then update status or delete status row.
func (p *Poller) processEvent(ctx context.Context, event Event) error {
	// Set status to in_progress.
	if err := p.updateStatus(ctx, event.ResourceType, event.ResourceID, "in_progress", ""); err != nil {
		p.logError(err, "failed to set in_progress status")
	}

	// Call the controller's handler.
	handleErr := p.handler.HandleEvent(ctx, event)

	if event.EventType == "deleted" {
		if handleErr == nil {
			// Success: delete the status row.
			return p.deleteStatus(ctx, event.ResourceType, event.ResourceID)
		}
		// Error on delete: set error status (don't delete the row).
		_ = p.updateStatus(ctx, event.ResourceType, event.ResourceID, "error", handleErr.Error())
		return handleErr
	}

	// Created event.
	if handleErr != nil {
		_ = p.updateStatus(ctx, event.ResourceType, event.ResourceID, "error", handleErr.Error())
		return handleErr
	}
	return p.updateStatus(ctx, event.ResourceType, event.ResourceID, "completed", "")
}

func (p *Poller) logError(err error, msg string) {
	if p.config.OnError != nil {
		p.config.OnError(err, msg)
	}
}

// --- HTTP helpers ---

func (p *Poller) doGet(ctx context.Context, url string) (*eventsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tenant manager returned %d: %s", resp.StatusCode, string(body))
	}

	var result eventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode events response: %w", err)
	}
	return &result, nil
}

func (p *Poller) updateStatus(ctx context.Context, resourceType string, resourceID uuid.UUID, status, message string) error {
	body := statusUpdateRequest{
		Controller:   p.controllerName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       status,
		Message:      message,
	}
	return p.doJSON(ctx, http.MethodPut, p.tenantManagerURL+"/v1/status", body)
}

func (p *Poller) deleteStatus(ctx context.Context, resourceType string, resourceID uuid.UUID) error {
	body := statusDeleteRequest{
		Controller:   p.controllerName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
	return p.doJSON(ctx, http.MethodDelete, p.tenantManagerURL+"/v1/status", body)
}

func (p *Poller) doJSON(ctx context.Context, method, url string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status API returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
