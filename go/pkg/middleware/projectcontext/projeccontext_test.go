// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package projectcontext

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bearerJWTWithProjectRole(t *testing.T, projectUUID string) string {
	t.Helper()

	headerJSON := []byte(`{"alg":"none","typ":"JWT"}`)
	payload, err := json.Marshal(map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": []interface{}{projectUUID + "_some-role"},
		},
	})
	require.NoError(t, err)

	// ParseUnverified only needs a well-formed JWT (3 segments). Signature can be empty.
	token := base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(payload) + "."
	return "Bearer " + token
}

func TestExtractProjectNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "valid project path",
			path:     "/v1/projects/someproject/resources",
			expected: "someproject",
		},
		{
			name:     "valid project path with hyphen, underscore and nubmers",
			path:     "/v1/projects/some-test123_project/some/nested/path",
			expected: "some-test123_project",
		},
		{
			name:     "old-path <-> no project",
			path:     "/edge-infra.orchestrator.apis/v2/resources",
			expected: "",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "root path",
			path:     "/",
			expected: "",
		},
		{
			name:     "malformed path - no trailing slash after project",
			path:     "/v1/projects/some-project",
			expected: "",
		},
		{
			name:     "path without /v1/projects prefix",
			path:     "/api/projects/some-project/resources",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractProjectNameFromPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveProjectUUID(t *testing.T) {
	tests := []struct {
		name           string
		projectName    string
		authHeader     string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
		expectedUUID   string
		errorContains  string
	}{
		{
			name:        "successful project resolution",
			projectName: "test-project",
			authHeader:  "Bearer token123",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]struct {
					Name   string `json:"name"`
					Status struct {
						ProjectStatus struct {
							UID string `json:"uID"`
						} `json:"projectStatus"`
					} `json:"status"`
				}{
					{
						Name: "test-project",
						Status: struct {
							ProjectStatus struct {
								UID string `json:"uID"`
							} `json:"projectStatus"`
						}{
							ProjectStatus: struct {
								UID string `json:"uID"`
							}{
								UID: "uuid-123",
							},
						},
					},
				})
			},
			expectError:  false,
			expectedUUID: "uuid-123",
		},
		{
			name:        "project not found",
			projectName: "non-existent",
			authHeader:  "Bearer token123",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]struct {
					Name   string `json:"name"`
					Status struct {
						ProjectStatus struct {
							UID string `json:"uID"`
						} `json:"projectStatus"`
					} `json:"status"`
				}{
					{
						Name: "other-project",
						Status: struct {
							ProjectStatus struct {
								UID string `json:"uID"`
							} `json:"projectStatus"`
						}{
							ProjectStatus: struct {
								UID string `json:"uID"`
							}{
								UID: "uuid-456",
							},
						},
					},
				})
			},
			expectError:   true,
			errorContains: "project not found: non-existent",
		},
		{
			name:        "API returns non-200 status",
			projectName: "test-project",
			authHeader:  "Bearer token123",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectError:   true,
			errorContains: "Nexus API returned status 500",
		},
		{
			name:        "invalid JSON response",
			projectName: "test-project",
			authHeader:  "Bearer token123",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			expectError:   true,
			errorContains: "failed to decode response",
		},
		{
			name:        "empty auth header",
			projectName: "test-project",
			authHeader:  "",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Empty(t, r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]struct {
					Name   string `json:"name"`
					Status struct {
						ProjectStatus struct {
							UID string `json:"uID"`
						} `json:"projectStatus"`
					} `json:"status"`
				}{})
			},
			expectError:   true,
			errorContains: "project not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			uuid, err := ResolveProjectUUID(context.Background(), tt.projectName, tt.authHeader, server.URL)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUUID, uuid)
			}
		})
	}
}

func TestResolveAndValidateProjectID(t *testing.T) {
	tests := []struct {
		name               string
		path               string
		authHeader         string
		existingProjectID  string
		config             ProjectResolverConfig
		setupServer        func() *httptest.Server
		mockValidateAccess func(authHeader, projectUUID string) error
		expectedUUID       string
		expectError        bool
		errorContains      string
	}{
		{
			name:              "returns existing project ID",
			path:              "/v1/projects/test/resources",
			authHeader:        "Bearer token",
			existingProjectID: "existing-uuid",
			config: ProjectResolverConfig{
				ProjectServiceURL:     "http://localhost",
				ErrorOnMissingProject: false,
			},
			expectedUUID: "existing-uuid",
			expectError:  false,
		},
		{
			name:              "old-style path extracts from JWT",
			path:              "/edge-infra.orchestrator.apis/v2/resources",
			authHeader:        bearerJWTWithProjectRole(t, "22222222-2222-2222-2222-222222222222"),
			existingProjectID: "",
			config: ProjectResolverConfig{
				ProjectServiceURL:     "http://localhost",
				ErrorOnMissingProject: false,
			},
			expectedUUID: "22222222-2222-2222-2222-222222222222",
			expectError:  false,
		},
		{
			name:              "no project name and no auth header",
			path:              "/some/other/path",
			authHeader:        "",
			existingProjectID: "",
			config: ProjectResolverConfig{
				ProjectServiceURL:     "http://localhost",
				ErrorOnMissingProject: false,
			},
			expectedUUID: "",
			expectError:  false,
		},
		{
			name:              "error on missing project when enabled",
			path:              "/v1/projects/non-existent/resources",
			authHeader:        "Bearer token",
			existingProjectID: "",
			config: ProjectResolverConfig{
				ErrorOnMissingProject: true,
			},
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]struct {
						Name   string `json:"name"`
						Status struct {
							ProjectStatus struct {
								UID string `json:"uID"`
							} `json:"projectStatus"`
						} `json:"status"`
					}{})
				}))
			},
			expectError:   true,
			errorContains: "failed to resolve project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.setupServer != nil {
				server = tt.setupServer()
				defer server.Close()
				tt.config.ProjectServiceURL = server.URL
			}

			uuid, err := ResolveAndValidateProjectID(
				context.Background(),
				tt.path,
				tt.authHeader,
				tt.existingProjectID,
				tt.config,
			)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedUUID, uuid)
			}
		})
	}
}

