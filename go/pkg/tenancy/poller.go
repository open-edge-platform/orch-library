// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
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
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/open-edge-platform/orch-library/go/dazl"
)

var log = dazl.GetLogger()

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

	// Timeout for individual HTTP requests (default 30s).
	HTTPTimeout time.Duration

	// OnError is an optional callback invoked when a non-fatal error occurs
	// (poll failure, event processing error, status update failure).
	OnError func(err error, msg string)
}

// DefaultPollerConfig returns sensible defaults.
func DefaultPollerConfig() PollerConfig {
	return PollerConfig{
		PollInterval:   5 * time.Second,
		PollLimit:      100,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		HTTPTimeout:    30 * time.Second,
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
// The internal /v1/events and /v1/status endpoints require no auth token —
// they are ClusterIP-only and the Tenant Manager enforces in-cluster network
// policy rather than JWT validation on these routes.
//
// Returns an error if tenantManagerURL or controllerName is empty, or if
// the resulting config is invalid.
func NewPoller(tenantManagerURL, controllerName string, handler Handler, opts ...func(*PollerConfig)) (*Poller, error) {
	if tenantManagerURL == "" {
		return nil, fmt.Errorf("tenantManagerURL must not be empty")
	}
	if controllerName == "" {
		return nil, fmt.Errorf("controllerName must not be empty")
	}
	if handler == nil {
		return nil, fmt.Errorf("handler must not be nil")
	}
	cfg := DefaultPollerConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	// Validate config after applying options.
	if cfg.PollInterval <= 0 {
		return nil, fmt.Errorf("PollInterval must be positive, got %s", cfg.PollInterval)
	}
	if cfg.PollLimit <= 0 {
		return nil, fmt.Errorf("PollLimit must be positive, got %d", cfg.PollLimit)
	}
	if cfg.InitialBackoff <= 0 {
		return nil, fmt.Errorf("InitialBackoff must be positive, got %s", cfg.InitialBackoff)
	}
	if cfg.MaxBackoff <= 0 {
		return nil, fmt.Errorf("MaxBackoff must be positive, got %s", cfg.MaxBackoff)
	}
	if cfg.HTTPTimeout <= 0 {
		return nil, fmt.Errorf("HTTPTimeout must be positive, got %s", cfg.HTTPTimeout)
	}

	return &Poller{
		tenantManagerURL: tenantManagerURL,
		controllerName:   controllerName,
		handler:          handler,
		config:           cfg,
		client:           &http.Client{Timeout: cfg.HTTPTimeout},
	}, nil
}

// Run executes replay (Phase 1), then enters steady-state polling
// (Phase 2). Blocks until ctx is cancelled. On restart, replays from
// scratch. Handlers must be idempotent.
func (p *Poller) Run(ctx context.Context) error {
	log.Infof("tenancy poller starting: controller=%s url=%s", p.controllerName, p.tenantManagerURL)

	// Phase 1: Replay with backoff retry.
	lastEventID, err := p.replayWithRetry(ctx)
	if err != nil {
		return err
	}

	log.Infof("tenancy poller replay complete: controller=%s lastEventId=%d", p.controllerName, lastEventID)

	// Phase 2: Steady-state polling.
	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Infof("tenancy poller stopping: controller=%s", p.controllerName)
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

		p.logError(err, fmt.Sprintf("replay failed (controller=%s), retrying in %s", p.controllerName, backoff))

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
// Returns an error if any event fails — the caller (replayWithRetry) will
// retry the whole replay until all synthesized events are successfully processed.
// Unlike poll(), partial progress is not tracked: the replay endpoint
// re-synthesizes the full snapshot on each call, so retrying from scratch is safe.
func (p *Poller) replay(ctx context.Context) (int64, error) {
	tmURL := fmt.Sprintf("%s/v1/events?controller=%s&replay=true",
		p.tenantManagerURL, url.QueryEscape(p.controllerName))

	resp, err := p.doGet(ctx, tmURL)
	if err != nil {
		return 0, err
	}

	log.Debugf("tenancy replay: controller=%s events=%d lastEventId=%d",
		p.controllerName, len(resp.Events), resp.LastEventID)

	for _, event := range resp.Events {
		if err := p.processEvent(ctx, event); err != nil {
			return 0, fmt.Errorf("replay event %s %s/%s failed (controller=%s): %w",
				event.EventType, event.ResourceType, event.ResourceName, p.controllerName, err)
		}
	}

	return resp.LastEventID, nil
}

// poll fetches incremental events after lastEventID and processes them.
// Returns the last successfully processed event ID. If any event fails,
// stops processing and returns the ID of the last successful event so that
// failed events will be retried on the next poll (at-least-once semantics).
func (p *Poller) poll(ctx context.Context, lastEventID int64) (int64, error) {
	tmURL := fmt.Sprintf("%s/v1/events?controller=%s&after=%d&limit=%d",
		p.tenantManagerURL, url.QueryEscape(p.controllerName), lastEventID, p.config.PollLimit)

	resp, err := p.doGet(ctx, tmURL)
	if err != nil {
		return lastEventID, err
	}

	if len(resp.Events) > 0 {
		log.Debugf("tenancy poll: controller=%s new_events=%d lastEventId=%d",
			p.controllerName, len(resp.Events), resp.LastEventID)
	}

	// Process events sequentially. Stop at first error to preserve
	// at-least-once semantics (failed event will be retried on next poll).
	processedLastEventID := lastEventID
	for _, event := range resp.Events {
		if err := p.processEvent(ctx, event); err != nil {
			p.logError(err, fmt.Sprintf("event %s %s/%s failed (controller=%s), will retry on next poll",
				event.EventType, event.ResourceType, event.ResourceName, p.controllerName))
			return processedLastEventID, nil
		}
		processedLastEventID = event.ID
	}

	return processedLastEventID, nil
}

// processEvent handles a single event: set in_progress, call handler,
// then update status or delete status row.
//
// Status update failures are logged but do NOT cause event processing to fail.
// This prevents temporary Tenant Manager unavailability from blocking event
// processing. The tradeoff: status may briefly be stale until the next successful
// status update (on replay or next event). Controllers must be idempotent to
// handle replay reconciliation.
func (p *Poller) processEvent(ctx context.Context, event Event) error {
	log.Debugf("tenancy processEvent: controller=%s type=%s/%s resource=%s id=%d",
		p.controllerName, event.ResourceType, event.EventType, event.ResourceName, event.ID)

	// Set status to in_progress. Log failure but continue processing.
	if err := p.updateStatus(ctx, event.ResourceType, event.ResourceID, StatusInProgress, ""); err != nil {
		p.logError(err, fmt.Sprintf("failed to set in_progress status for %s/%s (controller=%s)",
			event.ResourceType, event.ResourceName, p.controllerName))
	}

	// Call the controller's handler, wrapping any error with event context.
	handleErr := p.handler.HandleEvent(ctx, event)
	if handleErr != nil {
		handleErr = fmt.Errorf("handle %s %s/%s: %w",
			event.EventType, event.ResourceType, event.ResourceName, handleErr)
	}

	if event.EventType == EventTypeDeleted {
		if handleErr == nil {
			// Success: delete the status row (analogous to DeleteActiveWatchers in Nexus).
			// Status delete failure is logged but doesn't fail the event.
			if err := p.deleteStatus(ctx, event.ResourceType, event.ResourceID); err != nil {
				p.logError(err, fmt.Sprintf("failed to delete status for %s/%s (controller=%s)",
					event.ResourceType, event.ResourceName, p.controllerName))
			}
			return nil
		}
		// Error on delete: set error status (row remains for visibility).
		// Status update failure is logged but we still return the handler error.
		if err := p.updateStatus(ctx, event.ResourceType, event.ResourceID, StatusError, handleErr.Error()); err != nil {
			p.logError(err, fmt.Sprintf("failed to set error status for %s/%s (controller=%s)",
				event.ResourceType, event.ResourceName, p.controllerName))
		}
		return handleErr
	}

	// Created event.
	if handleErr != nil {
		if err := p.updateStatus(ctx, event.ResourceType, event.ResourceID, StatusError, handleErr.Error()); err != nil {
			p.logError(err, fmt.Sprintf("failed to set error status for %s/%s (controller=%s)",
				event.ResourceType, event.ResourceName, p.controllerName))
		}
		return handleErr
	}
	statusErr := p.updateStatus(ctx, event.ResourceType, event.ResourceID, StatusCompleted, "")
	if statusErr != nil {
		p.logError(statusErr, fmt.Sprintf("failed to set completed status for %s/%s (controller=%s)",
			event.ResourceType, event.ResourceName, p.controllerName))
	}
	return nil
}

func (p *Poller) logError(err error, msg string) {
	log.Warnf("tenancy poller: %s: %v", msg, err)
	if p.config.OnError != nil {
		p.config.OnError(err, msg)
	}
}

// --- HTTP helpers ---

func (p *Poller) doGet(ctx context.Context, reqURL string) (*eventsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
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

func (p *Poller) doJSON(ctx context.Context, method, reqURL string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(data))
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
