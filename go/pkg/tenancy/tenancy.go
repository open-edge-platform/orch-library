// SPDX-FileCopyrightText: 2026 Intel Corporation
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
	"time"

	"github.com/google/uuid"
)

// Event represents a tenancy lifecycle event. Both replay (synthesized
// from DB state) and incremental (from tenancy_events) events use the
// same structure.
type Event struct {
	ID           int64      `json:"id"`
	EventType    string     `json:"eventType"`    // "created", "deleted"
	ResourceType string     `json:"resourceType"` // "org", "project"
	ResourceID   uuid.UUID  `json:"resourceId"`
	ResourceName string     `json:"resourceName"`
	OrgID        *uuid.UUID `json:"orgId"`
	OrgName      *string    `json:"orgName"`
	FolderID     *uuid.UUID `json:"folderId"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// Handler is implemented by each controller with its business logic.
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
