// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"gotest.tools/assert"
)

func bearerJWTWithClaims(t *testing.T, claims map[string]interface{}) string {
	t.Helper()

	header := map[string]interface{}{
		"alg": "none",
		"typ": "JWT",
	}

	headerJSON, err := json.Marshal(header)
	assert.NilError(t, err)

	payloadJSON, err := json.Marshal(claims)
	assert.NilError(t, err)

	// ParseUnverified requires a token with 3 segments. Signature can be empty.
	token := base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(payloadJSON) + "."
	return "Bearer " + token
}

func bearerJWTWithRealmRoles(t *testing.T, roles []interface{}) string {
	t.Helper()

	return bearerJWTWithClaims(t, map[string]interface{}{
		"realm_access": map[string]interface{}{
			"roles": roles,
		},
	})
}

func TestExtractProjectIDFromJWT(t *testing.T) {
	validUUID1 := "11111111-1111-1111-1111-111111111111"
	validUUID2 := "22222222-2222-2222-2222-222222222222"

	// Note: These tests use ExtractProjectIDFromJWT which doesn't verify signatures.
	// The test JWTs are unsigned (alg: "none").
	// In production with upstream validation, use ExtractProjectIDFromJWT.
	// For standalone use without upstream validation, use ExtractProjectIDFromJWTWithVerification.

	tests := []struct {
		name            string
		authHeader      string
		expectedProject string
		expectedErrSub  string
		assertOneOf     []string
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedErrSub: "missing authorization header",
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "NotBearer token",
			expectedErrSub: "invalid authorization header format",
		},
		{
			name:           "malformed JWT",
			authHeader:     "Bearer not-a-jwt",
			expectedErrSub: "failed to parse JWT",
		},
		{
			name: "missing realm_access",
			authHeader: bearerJWTWithClaims(t, map[string]interface{}{
				"sub": "user",
			}),
			expectedErrSub: "realm_access not found",
		},
		{
			name: "missing roles",
			authHeader: bearerJWTWithClaims(t, map[string]interface{}{
				"realm_access": map[string]interface{}{},
			}),
			expectedErrSub: "roles not found",
		},
		{
			name: "roles wrong type",
			authHeader: bearerJWTWithClaims(t, map[string]interface{}{
				"realm_access": map[string]interface{}{
					"roles": "not-a-slice",
				},
			}),
			expectedErrSub: "roles not found",
		},
		{
			name: "no project ID found in roles",
			authHeader: bearerJWTWithRealmRoles(t, []interface{}{
				"admin",
				"viewer",
				"not-a-uuid_role",
			}),
			expectedErrSub: "no project ID found",
		},
		{
			name: "extracts project UUID from role",
			authHeader: bearerJWTWithRealmRoles(t, []interface{}{
				"admin",
				validUUID1 + "_some-role",
				"other",
			}),
			expectedProject: validUUID1,
		},
		{
			name: "multiple project UUIDs returns one of them",
			authHeader: bearerJWTWithRealmRoles(t, []interface{}{
				validUUID1 + "_role-a",
				validUUID2 + "_role-b",
			}),
			assertOneOf: []string{validUUID1, validUUID2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectID, err := ExtractProjectIDFromJWT(tt.authHeader)
			if tt.expectedErrSub != "" {
				assert.ErrorContains(t, err, tt.expectedErrSub)
				assert.Equal(t, "", projectID)
				return
			}
			assert.NilError(t, err)
			if len(tt.assertOneOf) > 0 {
				ok := false
				for _, candidate := range tt.assertOneOf {
					if projectID == candidate {
						ok = true
						break
					}
				}
				assert.Assert(t, ok)
				return
			}
			assert.Equal(t, tt.expectedProject, projectID)
		})
	}
}

func TestExtractAllProjectIDsFromJWT(t *testing.T) {
	validUUID1 := "11111111-1111-1111-1111-111111111111"
	validUUID2 := "22222222-2222-2222-2222-222222222222"

	tests := []struct {
		name           string
		authHeader     string
		expectedIDs    []string
		expectedErrSub string
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedErrSub: "missing authorization header",
		},
		{
			name:           "invalid authorization header format",
			authHeader:     "Bearer",
			expectedErrSub: "invalid authorization header format",
		},
		{
			name:           "malformed JWT",
			authHeader:     "Bearer not-a-jwt",
			expectedErrSub: "failed to parse JWT",
		},
		{
			name: "returns all UUIDs from roles",
			authHeader: bearerJWTWithRealmRoles(t, []interface{}{
				"admin",
				validUUID1 + "_some-role",
				validUUID2 + "_another-role",
				"not-a-uuid_role",
			}),
			expectedIDs: []string{validUUID1, validUUID2},
		},
		{
			name: "no project IDs found",
			authHeader: bearerJWTWithRealmRoles(t, []interface{}{
				"admin",
			}),
			expectedErrSub: "no project IDs found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectIDs, err := ExtractAllProjectIDsFromJWT(tt.authHeader)
			if tt.expectedErrSub != "" {
				assert.ErrorContains(t, err, tt.expectedErrSub)
				assert.Assert(t, projectIDs == nil)
				return
			}
			assert.NilError(t, err)
			for _, id := range tt.expectedIDs {
				assert.Assert(t, projectIDs[id])
			}
			assert.Assert(t, len(projectIDs) == len(tt.expectedIDs))
		})
	}
}

func TestValidateProjectAccess(t *testing.T) {
	validUUID := "11111111-1111-1111-1111-111111111111"
	otherUUID := "22222222-2222-2222-2222-222222222222"

	tests := []struct {
		name           string
		authHeader     string
		projectID      string
		expectedErrSub string
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			projectID:      validUUID,
			expectedErrSub: "missing authorization header",
		},
		{
			name:           "missing project ID",
			authHeader:     "Bearer token",
			projectID:      "",
			expectedErrSub: "missing project ID",
		},
		{
			name:           "propagates extraction error",
			authHeader:     "Bearer not-a-jwt",
			projectID:      validUUID,
			expectedErrSub: "failed to extract projects from JWT",
		},
		{
			name: "denies access when project not in JWT roles",
			authHeader: bearerJWTWithRealmRoles(t, []interface{}{
				otherUUID + "_some-role",
			}),
			projectID:      validUUID,
			expectedErrSub: "user does not have access",
		},
		{
			name: "allows access when project in JWT roles",
			authHeader: bearerJWTWithRealmRoles(t, []interface{}{
				validUUID + "_some-role",
			}),
			projectID: validUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectAccess(tt.authHeader, tt.projectID)
			if tt.expectedErrSub != "" {
				assert.ErrorContains(t, err, tt.expectedErrSub)
				return
			}
			assert.NilError(t, err)
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected bool
	}{
		{
			name:     "valid UUID",
			uuid:     "123e4567-e89b-12d3-a456-426614174000",
			expected: true,
		},
		{
			name:     "valid UUID uppercase",
			uuid:     "123E4567-E89B-12D3-A456-426614174000",
			expected: true,
		},
		{
			name:     "invalid UUID",
			uuid:     "not-a-uuid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isValidUUID(tt.uuid))
		})
	}
}
