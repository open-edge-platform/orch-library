// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0
package projectcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/open-edge-platform/orch-library/go/pkg/auth"
)

const (
	ActiveProjectIDHeader = "ActiveProjectID"
	pathProjectPattern    = `^/v1/projects/([^/]+)/`
)

var projectPathRegex = regexp.MustCompile(pathProjectPattern)

// ExtractProjectNameFromPath extracts the project name from the given context
func ExtractProjectNameFromPath(path string) string {
	matches := projectPathRegex.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ResolveProjectUUID queries the project service API to resolve project UUID from project name
func ResolveProjectUUID(ctx context.Context, projectName string, authHeader string, projectServiceURL string) (string, error) {
	reqURL := fmt.Sprintf("%s/v1/projects?member-role=true", projectServiceURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to query Nexus API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Nexus API returned status %d", resp.StatusCode) //nolint:staticcheck
	}

	var projects []struct {
		Name   string `json:"name"`
		Status struct {
			ProjectStatus struct {
				UID string `json:"uID"`
			} `json:"projectStatus"`
		} `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	for _, project := range projects {
		if project.Name == projectName {
			return project.Status.ProjectStatus.UID, nil
		}
	}

	return "", fmt.Errorf("project not found: %s", projectName)
}

// ProjectResolverConfig holds configuration for project resolution and validation
type ProjectResolverConfig struct {
	ProjectServiceURL     string
	ErrorOnMissingProject bool
}

// ResolveAndValidateProjectID is a framework-agnostic helper that resolves and validates
// project ID from request path and auth header. It performs:
// 1. Extract project name from path
// 2. Resolve project UUID via project service (Nexus) API
// 3. Validate user has access to the project
// 4. Fall back to JWT extraction for old-style paths
//
// Returns (projectUUID, error) - error is non-nil only if ErrorOnMissingProject is true
// and the project cannot be resolved or user doesn't have access.
func ResolveAndValidateProjectID(ctx context.Context, path string, authHeader string, existingProjectID string, config ProjectResolverConfig) (string, error) {
	if existingProjectID != "" {
		return existingProjectID, nil // Already set, no need to resolve
	}

	// Try to extract project name from URL path (new multi-tenant paths)
	projectName := ExtractProjectNameFromPath(path)

	if projectName != "" {
		// New-style path: /v1/projects/{projectName}/...
		projectUUID, err := ResolveProjectUUID(ctx, projectName, authHeader, config.ProjectServiceURL)
		if err != nil {
			if config.ErrorOnMissingProject {
				return "", fmt.Errorf("failed to resolve project: %w", err)
			}
			return "", nil
		}

		if projectUUID != "" {
			// Validate user has access to this project
			if err := auth.ValidateProjectAccess(authHeader, projectUUID); err != nil {
				if config.ErrorOnMissingProject {
					return "", fmt.Errorf("access denied to project %s: %w", projectName, err)
				}
				return "", nil
			}
			return projectUUID, nil
		}
	} else if authHeader != "" {
		// Old-style path: /edge-infra.orchestrator.apis/v2/...
		// Extract project UUID from JWT token roles
		projectUUID, _ := auth.ExtractProjectIDFromJWT(authHeader)
		return projectUUID, nil
	}

	return "", nil
}

// InjectActiveProjectID is a standard http.Handler middleware that resolves and injects
// the active project ID with security validation.
func InjectActiveProjectID(projectServiceURL string, errorOnMissing bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(ActiveProjectIDHeader) == "" {
				authHeader := r.Header.Get("Authorization")

				// Use the helper function with validation
				projectUUID, err := ResolveAndValidateProjectID(
					r.Context(),
					r.URL.Path,
					authHeader,
					"",
					ProjectResolverConfig{
						ProjectServiceURL:     projectServiceURL,
						ErrorOnMissingProject: errorOnMissing,
					},
				)

				if err != nil && errorOnMissing {
					// Return 403 for unauthorized access
					http.Error(w, "Access denied or project not found", http.StatusForbidden)
					return
				}

				if projectUUID != "" {
					r.Header.Set(ActiveProjectIDHeader, projectUUID)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
