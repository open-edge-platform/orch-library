// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package tenancy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingHandler records every event it receives and optionally returns an error.
type recordingHandler struct {
	events []Event
	errFn  func(Event) error
}

func (h *recordingHandler) HandleEvent(_ context.Context, e Event) error {
	h.events = append(h.events, e)
	if h.errFn != nil {
		return h.errFn(e)
	}
	return nil
}

// makeEventsServer builds a test HTTP server that serves event payloads and
// records PUT/DELETE /v1/status calls.
type statusCall struct {
	method string
	body   map[string]interface{}
}

type testServer struct {
	*httptest.Server
	statusCalls []statusCall
}

func newTestServer(t *testing.T, eventsPayloads map[string]eventsResponse) *testServer {
	t.Helper()
	ts := &testServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/events", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		payload, ok := eventsPayloads[key]
		if !ok {
			// Default: empty events, lastEventId=0
			payload = eventsResponse{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	})

	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		ts.statusCalls = append(ts.statusCalls, statusCall{method: r.Method, body: body})
		w.WriteHeader(http.StatusOK)
	})

	ts.Server = httptest.NewServer(mux)
	return ts
}

// mustPoller creates a Poller for tests, failing fast if construction fails.
func mustPoller(t *testing.T, serverURL, name string, h Handler, opts ...func(*PollerConfig)) *Poller {
	t.Helper()
	p, err := NewPoller(serverURL, name, h, opts...)
	require.NoError(t, err)
	return p
}

// ─── DefaultPollerConfig ────────────────────────────────────────────────────

func TestDefaultPollerConfig(t *testing.T) {
	cfg := DefaultPollerConfig()
	assert.Equal(t, 5*time.Second, cfg.PollInterval)
	assert.Equal(t, 100, cfg.PollLimit)
	assert.Equal(t, 1*time.Second, cfg.InitialBackoff)
	assert.Equal(t, 30*time.Second, cfg.MaxBackoff)
}

// ─── NewPoller validation ────────────────────────────────────────────────────

func TestNewPoller_AppliesOptions(t *testing.T) {
	h := &recordingHandler{}
	p, err := NewPoller("http://localhost", "my-ctrl", h, func(cfg *PollerConfig) {
		cfg.PollInterval = 42 * time.Second
		cfg.PollLimit = 7
	})
	require.NoError(t, err)
	assert.Equal(t, 42*time.Second, p.config.PollInterval)
	assert.Equal(t, 7, p.config.PollLimit)
	assert.Equal(t, "my-ctrl", p.controllerName)
}

func TestNewPoller_RejectsEmptyURL(t *testing.T) {
	_, err := NewPoller("", "my-ctrl", &recordingHandler{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenantManagerURL")
}

func TestNewPoller_RejectsEmptyControllerName(t *testing.T) {
	_, err := NewPoller("http://localhost", "", &recordingHandler{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "controllerName")
}

func TestNewPoller_RejectsNilHandler(t *testing.T) {
	_, err := NewPoller("http://localhost", "my-ctrl", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler")
}

func TestNewPoller_ValidatesConfig(t *testing.T) {
	h := &recordingHandler{}

	tests := []struct {
		name      string
		opt       func(*PollerConfig)
		expectErr string
	}{
		{
			name: "negative poll interval",
			opt: func(c *PollerConfig) {
				c.PollInterval = -1 * time.Second
			},
			expectErr: "PollInterval must be positive",
		},
		{
			name: "zero poll limit",
			opt: func(c *PollerConfig) {
				c.PollLimit = 0
			},
			expectErr: "PollLimit must be positive",
		},
		{
			name: "negative initial backoff",
			opt: func(c *PollerConfig) {
				c.InitialBackoff = -1 * time.Second
			},
			expectErr: "InitialBackoff must be positive",
		},
		{
			name: "zero max backoff",
			opt: func(c *PollerConfig) {
				c.MaxBackoff = 0
			},
			expectErr: "MaxBackoff must be positive",
		},
		{
			name: "negative http timeout",
			opt: func(c *PollerConfig) {
				c.HTTPTimeout = -5 * time.Second
			},
			expectErr: "HTTPTimeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPoller("http://tm", "ctrl", h, tt.opt)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectErr)
		})
	}
}

// ─── replay: happy path ──────────────────────────────────────────────────────

func TestPoller_Replay_ProcessesEvents(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	orgID := uuid.New()
	orgName := "my-org"

	replayResp := eventsResponse{
		LastEventID: 2,
		Events: []Event{
			{
				ID: 1, EventType: EventTypeCreated, ResourceType: ResourceTypeOrg,
				ResourceID: id1, ResourceName: "my-org", CreatedAt: time.Now(),
			},
			{
				ID: 2, EventType: EventTypeCreated, ResourceType: ResourceTypeProject,
				ResourceID: id2, ResourceName: "my-project",
				OrgID: &orgID, OrgName: &orgName, CreatedAt: time.Now(),
			},
		},
	}

	srv := newTestServer(t, map[string]eventsResponse{
		"controller=test-ctrl&replay=true": replayResp,
	})
	defer srv.Close()

	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "test-ctrl", h)

	lastID, err := p.replay(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), lastID)
	require.Len(t, h.events, 2)
	assert.Equal(t, "my-org", h.events[0].ResourceName)
	assert.Equal(t, "my-project", h.events[1].ResourceName)
}

// ─── replay: status updates via processEvent ─────────────────────────────────

func TestPoller_ProcessEvent_CreatedSuccess(t *testing.T) {
	id := uuid.New()
	event := Event{
		ID: 1, EventType: EventTypeCreated, ResourceType: ResourceTypeProject,
		ResourceID: id, ResourceName: "proj",
	}

	var statusCalls []statusCall
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		statusCalls = append(statusCalls, statusCall{method: r.Method, body: body})
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "ctrl", h)

	err := p.processEvent(context.Background(), event)
	require.NoError(t, err)

	require.Len(t, statusCalls, 2)
	assert.Equal(t, http.MethodPut, statusCalls[0].method)
	assert.Equal(t, StatusInProgress, statusCalls[0].body["status"])
	assert.Equal(t, http.MethodPut, statusCalls[1].method)
	assert.Equal(t, StatusCompleted, statusCalls[1].body["status"])
}

func TestPoller_ProcessEvent_CreatedError(t *testing.T) {
	id := uuid.New()
	event := Event{
		ID: 1, EventType: EventTypeCreated, ResourceType: ResourceTypeProject,
		ResourceID: id, ResourceName: "proj",
	}

	var statusCalls []statusCall
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		statusCalls = append(statusCalls, statusCall{method: r.Method, body: body})
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	h := &recordingHandler{errFn: func(Event) error { return errors.New("handler failed") }}
	p := mustPoller(t, srv.URL, "ctrl", h)

	err := p.processEvent(context.Background(), event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler failed")
	// Error is wrapped with event context.
	assert.Contains(t, err.Error(), "proj")

	require.Len(t, statusCalls, 2)
	assert.Equal(t, StatusInProgress, statusCalls[0].body["status"])
	assert.Equal(t, StatusError, statusCalls[1].body["status"])
}

func TestPoller_ProcessEvent_DeletedSuccess(t *testing.T) {
	id := uuid.New()
	event := Event{
		ID: 1, EventType: EventTypeDeleted, ResourceType: ResourceTypeProject,
		ResourceID: id, ResourceName: "proj",
	}

	var statusCalls []statusCall
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		statusCalls = append(statusCalls, statusCall{method: r.Method, body: body})
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "ctrl", h)

	err := p.processEvent(context.Background(), event)
	require.NoError(t, err)

	require.Len(t, statusCalls, 2)
	// First call: in_progress (PUT)
	assert.Equal(t, http.MethodPut, statusCalls[0].method)
	assert.Equal(t, StatusInProgress, statusCalls[0].body["status"])
	// Second call: DELETE (no status field)
	assert.Equal(t, http.MethodDelete, statusCalls[1].method)
}

func TestPoller_ProcessEvent_DeletedError(t *testing.T) {
	id := uuid.New()
	event := Event{
		ID: 1, EventType: EventTypeDeleted, ResourceType: ResourceTypeProject,
		ResourceID: id, ResourceName: "proj",
	}

	var statusCalls []statusCall
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		statusCalls = append(statusCalls, statusCall{method: r.Method, body: body})
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	h := &recordingHandler{errFn: func(Event) error { return errors.New("cleanup failed") }}
	p := mustPoller(t, srv.URL, "ctrl", h)

	err := p.processEvent(context.Background(), event)
	require.Error(t, err)

	require.Len(t, statusCalls, 2)
	assert.Equal(t, StatusInProgress, statusCalls[0].body["status"])
	// On delete error: set error status (PUT, not DELETE)
	assert.Equal(t, http.MethodPut, statusCalls[1].method)
	assert.Equal(t, StatusError, statusCalls[1].body["status"])
}

// ─── poll: incremental events ───────────────────────────────────────────────

func TestPoller_Poll_AdvancesLastEventID(t *testing.T) {
	id := uuid.New()
	incrementalResp := eventsResponse{
		LastEventID: 5,
		Events: []Event{
			{ID: 5, EventType: EventTypeCreated, ResourceType: ResourceTypeOrg, ResourceID: id, ResourceName: "org2"},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/events", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(incrementalResp)
	})
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "ctrl", h)

	newID, err := p.poll(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, int64(5), newID)
	require.Len(t, h.events, 1)
	assert.Equal(t, "org2", h.events[0].ResourceName)
}

func TestPoller_Poll_PreservesLastIDOnError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/events", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "ctrl", h)

	newID, err := p.poll(context.Background(), 42)
	require.Error(t, err)
	assert.Equal(t, int64(42), newID) // unchanged on error
}

// ─── replayWithRetry: backoff ────────────────────────────────────────────────

func TestPoller_ReplayWithRetry_SucceedsAfterFailures(t *testing.T) {
	var callCount atomic.Int32

	id := uuid.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/events", func(w http.ResponseWriter, _ *http.Request) {
		n := callCount.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(eventsResponse{
			LastEventID: 10,
			Events: []Event{
				{ID: 10, EventType: EventTypeCreated, ResourceType: ResourceTypeOrg, ResourceID: id, ResourceName: "org1"},
			},
		})
	})
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	var errMsgs []string
	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "ctrl", h, func(cfg *PollerConfig) {
		cfg.InitialBackoff = 1 * time.Millisecond
		cfg.MaxBackoff = 5 * time.Millisecond
		cfg.OnError = func(_ error, msg string) { errMsgs = append(errMsgs, msg) }
	})

	lastID, err := p.replayWithRetry(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(10), lastID)
	assert.GreaterOrEqual(t, int(callCount.Load()), 3)
	assert.NotEmpty(t, errMsgs) // errors were reported during retries
}

func TestPoller_ReplayWithRetry_CancelledContext(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/events", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "ctrl", h, func(cfg *PollerConfig) {
		cfg.InitialBackoff = 5 * time.Millisecond
		cfg.MaxBackoff = 10 * time.Millisecond
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_, err := p.replayWithRetry(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))
}

// ─── Run: end-to-end ─────────────────────────────────────────────────────────

func TestPoller_Run_CancelsCleanly(t *testing.T) {
	id := uuid.New()
	replayResp := eventsResponse{
		LastEventID: 1,
		Events: []Event{
			{ID: 1, EventType: EventTypeCreated, ResourceType: ResourceTypeOrg, ResourceID: id, ResourceName: "org1"},
		},
	}

	var pollCount atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/events", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("replay") == "true" {
			_ = json.NewEncoder(w).Encode(replayResp)
			return
		}
		pollCount.Add(1)
		_ = json.NewEncoder(w).Encode(eventsResponse{LastEventID: 1})
	})
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "ctrl", h, func(cfg *PollerConfig) {
		cfg.PollInterval = 5 * time.Millisecond
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := p.Run(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled))

	// Replay processed the one event; steady-state poll ran at least once.
	require.Len(t, h.events, 1)
	assert.GreaterOrEqual(t, int(pollCount.Load()), 1)
}

// ─── onError callback ────────────────────────────────────────────────────────

func TestPoller_OnError_CalledOnPollFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/events", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("replay") == "true" {
			_ = json.NewEncoder(w).Encode(eventsResponse{LastEventID: 0})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "internal error")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	var errorMessages []string
	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "ctrl", h, func(cfg *PollerConfig) {
		cfg.PollInterval = 5 * time.Millisecond
		cfg.OnError = func(_ error, msg string) {
			errorMessages = append(errorMessages, msg)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	_ = p.Run(ctx)
	require.NotEmpty(t, errorMessages)
	assert.Contains(t, errorMessages[0], "poll failed")
}

// ─── status request body correctness ────────────────────────────────────────

func TestPoller_StatusRequest_ContainsCorrectFields(t *testing.T) {
	projectID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	event := Event{
		ID: 99, EventType: EventTypeCreated, ResourceType: ResourceTypeProject,
		ResourceID: projectID, ResourceName: "my-proj",
	}

	var puts []map[string]interface{}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			puts = append(puts, body)
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	h := &recordingHandler{}
	p := mustPoller(t, srv.URL, "my-controller", h)
	_ = p.processEvent(context.Background(), event)

	require.Len(t, puts, 2)
	for _, put := range puts {
		assert.Equal(t, "my-controller", put["controller"])
		assert.Equal(t, ResourceTypeProject, put["resourceType"])
		assert.Equal(t, projectID.String(), put["resourceId"])
	}
	assert.Equal(t, StatusInProgress, puts[0]["status"])
	assert.Equal(t, StatusCompleted, puts[1]["status"])
}

// ─── Event.String() ──────────────────────────────────────────────────────────

func TestEvent_String(t *testing.T) {
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	e := Event{
		ID:           42,
		EventType:    EventTypeCreated,
		ResourceType: ResourceTypeProject,
		ResourceID:   id,
		ResourceName: "my-proj",
	}
	s := e.String()
	assert.Contains(t, s, "42")
	assert.Contains(t, s, EventTypeCreated)
	assert.Contains(t, s, ResourceTypeProject)
	assert.Contains(t, s, "my-proj")
	assert.Contains(t, s, id.String())
}
