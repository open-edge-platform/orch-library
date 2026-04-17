// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package tenancy provides a shared client library for tenant controllers
// to consume tenancy events from the Tenant Manager REST API.
//
// This package has no database or Ent dependencies -- it is a pure HTTP
// client. Controllers only need the Tenant Manager URL and their canonical
// controller name.
package tenancy

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EventType constants for Event.EventType.
const (
	EventTypeCreated = "created"
	EventTypeDeleted = "deleted"
)

// ResourceType constants for Event.ResourceType.
const (
	ResourceTypeOrg     = "org"
	ResourceTypeProject = "project"
)

// Status constants used in controller status reporting.
const (
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusError      = "error"
)

// Event represents a tenancy lifecycle event. Both replay (synthesized
// from DB state) and incremental (from tenancy_events) events use the
// same structure. Handlers must be idempotent — replay on restart will
// re-deliver events for all existing and soft-deleted resources.
type Event struct {
	// ID is the monotonically increasing event sequence number.
	ID int64 `json:"id"`
	// EventType is "created" or "deleted".
	EventType string `json:"eventType"`
	// ResourceType is "org" or "project".
	ResourceType string    `json:"resourceType"`
	ResourceID   uuid.UUID `json:"resourceId"`
	ResourceName string    `json:"resourceName"`
	// OrgID is set for project events; nil for org events.
	OrgID   *uuid.UUID `json:"orgId"`
	OrgName *string    `json:"orgName"`
	// FolderID is reserved for future hierarchical resource types.
	FolderID  *uuid.UUID `json:"folderId"`
	CreatedAt time.Time  `json:"createdAt"`
}

// String returns a summary of the event for logging.
func (e Event) String() string {
	return fmt.Sprintf("Event{id=%d type=%s/%s name=%q resourceId=%s}",
		e.ID, e.ResourceType, e.EventType, e.ResourceName, e.ResourceID)
}

type Handler interface {
	// HandleEvent is called for each event (both replay and incremental).
	// Must be idempotent -- replay on restart will re-deliver events for
	// all existing and soft-deleted resources.
	HandleEvent(ctx context.Context, event Event) error
}

// eventsResponse is the wire format returned by GET /v1/events.
type eventsResponse struct {
	Events      []Event `json:"events"`
	LastEventID int64   `json:"lastEventId"`
}

// statusUpdateRequest is the body for PUT /v1/status.
type statusUpdateRequest struct {
	Controller   string    `json:"controller"`
	ResourceType string    `json:"resourceType"`
	ResourceID   uuid.UUID `json:"resourceId"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
}

// statusDeleteRequest is the body for DELETE /v1/status.
type statusDeleteRequest struct {
	Controller   string    `json:"controller"`
	ResourceType string    `json:"resourceType"`
	ResourceID   uuid.UUID `json:"resourceId"`
}
