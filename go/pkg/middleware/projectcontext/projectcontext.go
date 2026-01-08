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

const ActiveProjectIDHeader = "ActiveProjectID"

var projectPathRegex = regexp.MustCompile(`^/v1/projects/([^/]+)/`)

// ExtractProjectNameFromPath extracts the project name from the given context
func ExtractProjectNameFromPath(path string) string {
	matches := projectPathRegex.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ResolveProjectUUID queries the Nexus API to resolve project UUID from project name
func ResolveProjectUUID(ctx context.Context, projectName string, authHeader string, nexusAPIURL string) (string, error) {
	reqURL := fmt.Sprintf("%s/v1/projects", nexusAPIURL)

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
		return "", fmt.Errorf("Nexus API returned status %d", resp.StatusCode)
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

// InjectActiveProjectID is a middleware that resolves and injects the active project ID
// This can be used with Echo or other HTTP frameworks
func InjectActiveProjectID(nexusAPIURL string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(ActiveProjectIDHeader) == "" {
			authHeader := r.Header.Get("Authorization")

			// Try to extract project name from URL path
			projectName := ExtractProjectNameFromPath(r.URL.Path)
			if projectName != "" {
				// New-style path: /v1/projects/{projectName}/...
				projectUUID, err := ResolveProjectUUID(r.Context(), projectName, authHeader, nexusAPIURL)
				if err == nil && projectUUID != "" {
					r.Header.Set(ActiveProjectIDHeader, projectUUID)
				}
			} else if authHeader != "" {
				// Old-style path: extract from JWT
				projectUUID, err := auth.ExtractProjectIDFromJWT(authHeader)
				if err == nil && projectUUID != "" {
					r.Header.Set(ActiveProjectIDHeader, projectUUID)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
