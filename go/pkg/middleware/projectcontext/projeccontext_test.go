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

func TestInjectActiveProjectID(t *testing.T) {
	tests := []struct {
		name              string
		nexusAPIURL       string
		errorOnMissing    bool
		existingHeader    string
		requestPath       string
		authHeader        string
		setupServer       func() *httptest.Server
		expectedHeader    string
		expectedStatus    int
		expectHandlerCall bool
	}{
		{
			name:              "skips resolution when header already set",
			nexusAPIURL:       "http://localhost",
			errorOnMissing:    false,
			existingHeader:    "existing-uuid",
			requestPath:       "/v1/projects/test/resources",
			authHeader:        "Bearer token",
			expectedHeader:    "existing-uuid",
			expectedStatus:    http.StatusOK,
			expectHandlerCall: true,
		},
		{
			name:           "sets header on successful resolution",
			errorOnMissing: false,
			existingHeader: "",
			requestPath:    "/v1/projects/test-project/resources",
			authHeader:     bearerJWTWithProjectRole(t, "11111111-1111-1111-1111-111111111111"),
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
									UID: "11111111-1111-1111-1111-111111111111",
								},
							},
						},
					})
				}))
			},

			expectedHeader:    "11111111-1111-1111-1111-111111111111",
			expectedStatus:    http.StatusOK,
			expectHandlerCall: true,
		},
		{
			name: "continues without header when error on missing disabled",
			// for old-style path + missing Authorization header, errorOnMissing=false
            // the middleware must not block the request and must NOT inject ActiveProjectID,
            // since it cannot resolve one.
			errorOnMissing:    false,
			existingHeader:    "",
			requestPath:       "/edge-infra.orchestrator.apis/v2/hosts",
			authHeader:        "",
			expectedHeader:    "",
			expectedStatus:    http.StatusOK,
			expectHandlerCall: true,
		},
		{
			name:              "sets header from JWT on old-style path",
			errorOnMissing:    false,
			existingHeader:    "",
			requestPath:       "/edge-infra.orchestrator.apis/v2/hosts",
			authHeader:        bearerJWTWithProjectRole(t, "33333333-3333-3333-3333-333333333333"),
			expectedHeader:    "33333333-3333-3333-3333-333333333333",
			expectedStatus:    http.StatusOK,
			expectHandlerCall: true,
		},
		{
			name:           "returns 403 when access validation fails in strict mode",
			errorOnMissing: true,
			existingHeader: "",
			requestPath:    "/v1/projects/test-project/resources",
			// JWT contains a *different* project UUID than the one resolved from the project service.
			// This should cause auth.ValidateProjectAccess to fail.
			authHeader: bearerJWTWithProjectRole(t, "44444444-4444-4444-4444-444444444444"),
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
									UID: "55555555-5555-5555-5555-555555555555",
								},
							},
						},
					})
				}))
			},
			expectedHeader:    "",
			expectedStatus:    http.StatusForbidden,
			expectHandlerCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nexusURL := tt.nexusAPIURL
			if tt.setupServer != nil {
				server := tt.setupServer()
				defer server.Close()
				nexusURL = server.URL
			}

			handlerCalled := false
			var actualHeader string
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				actualHeader = r.Header.Get(ActiveProjectIDHeader)
				w.WriteHeader(http.StatusOK)
			})

			handler := InjectActiveProjectID(nexusURL, tt.errorOnMissing)(nextHandler)

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			if tt.existingHeader != "" {
				req.Header.Set(ActiveProjectIDHeader, tt.existingHeader)
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectHandlerCall, handlerCalled)

			assert.Equal(t, tt.expectedHeader, actualHeader)
		})
	}
}
