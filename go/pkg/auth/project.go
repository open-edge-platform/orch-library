// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0
package auth

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Project represents a project within the authentication system - the tenant-aware authentication

const (
	roleProjectIDSeparator = "_"
	uuidPattern            = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
)

var uuidRegex = regexp.MustCompile(uuidPattern)

// parseToken parses a JWT token string with optional signature verification.
// When verifySignature is true, it validates the token using the ORCH_JWT_SIGNING_KEY.
// When verifySignature is false, it parses without verification (use only when JWT was already validated upstream).
func parseToken(tokenString string, verifySignature bool) (jwt.MapClaims, error) {
	if verifySignature {
		secret := os.Getenv("ORCH_JWT_SIGNING_KEY")
		if secret == "" {
			return nil, fmt.Errorf("JWT signing key not configured")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Reject tokens using the "none" signing method or unexpected algorithms.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to parse JWT: %w", err)
		}

		if !token.Valid {
			return nil, fmt.Errorf("invalid JWT token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("invalid JWT claims type")
		}

		return claims, nil
	}

	// Parse without verification - use only when JWT was already validated upstream
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	return claims, nil
}

// ExtractProjectIDFromJWT extracts the project UUID from JWT token roles without signature verification.
// This function assumes the JWT has already been validated by upstream authentication middleware.
// Use ExtractProjectIDFromJWTWithVerification if you need signature verification.
func ExtractProjectIDFromJWT(authHeader string) (string, error) {
	return extractProjectIDFromJWT(authHeader, false)
}

// ExtractProjectIDFromJWTWithVerification extracts the project UUID from JWT token roles
// and verifies the JWT signature using ORCH_JWT_SIGNING_KEY.
// Use this when the JWT has NOT been validated by upstream middleware.
func ExtractProjectIDFromJWTWithVerification(authHeader string) (string, error) {
	return extractProjectIDFromJWT(authHeader, true)
}

// extractProjectIDFromJWT is the internal implementation that handles both verified and unverified parsing.
func extractProjectIDFromJWT(authHeader string, verifySignature bool) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]

	// Parse and validate token to extract trusted claims
	claims, err := parseToken(tokenString, verifySignature)
	if err != nil {
		return "", err
	}

	// Extract roles from realm_access.roles
	realmAccess, ok := claims["realm_access"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("realm_access not found in JWT")
	}

	rolesInterface, ok := realmAccess["roles"].([]interface{})
	if !ok {
		return "", fmt.Errorf("roles not found in realm_access")
	}

	// Extract project UUIDs from roles
	projectIDs := make(map[string]bool)
	for _, roleInterface := range rolesInterface {
		role, ok := roleInterface.(string)
		if !ok {
			continue
		}

		// Roles with project context follow pattern: {projectUUID}_{roleName}
		if strings.Contains(role, roleProjectIDSeparator) {
			parts := strings.Split(role, roleProjectIDSeparator)
			if len(parts) > 0 {
				potentialUUID := parts[0]
				if uuidRegex.MatchString(potentialUUID) {
					projectIDs[potentialUUID] = true
				}
			}
		}
	}

	if len(projectIDs) == 0 {
		return "", fmt.Errorf("no project ID found in JWT roles")
	}

	// If user has multiple projects, use the first one found
	var projectID string
	for id := range projectIDs {
		projectID = id
		break
	}

	return projectID, nil
}

// ExtractAllProjectIDsFromJWT extracts all project UUIDs from JWT token roles without signature verification.
// Returns a map of project IDs for quick lookup.
// This function assumes the JWT has already been validated by upstream authentication middleware.
// Use ExtractAllProjectIDsFromJWTWithVerification if you need signature verification.
func ExtractAllProjectIDsFromJWT(authHeader string) (map[string]bool, error) {
	return extractAllProjectIDsFromJWT(authHeader, false)
}

// ExtractAllProjectIDsFromJWTWithVerification extracts all project UUIDs from JWT token roles
// and verifies the JWT signature using ORCH_JWT_SIGNING_KEY.
// Use this when the JWT has NOT been validated by upstream middleware.
func ExtractAllProjectIDsFromJWTWithVerification(authHeader string) (map[string]bool, error) {
	return extractAllProjectIDsFromJWT(authHeader, true)
}

// extractAllProjectIDsFromJWT is the internal implementation that handles both verified and unverified parsing.
func extractAllProjectIDsFromJWT(authHeader string, verifySignature bool) (map[string]bool, error) {
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]

	// Parse and validate token to extract trusted claims
	claims, err := parseToken(tokenString, verifySignature)
	if err != nil {
		return nil, err
	}

	// Extract roles from realm_access.roles
	realmAccess, ok := claims["realm_access"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("realm_access not found in JWT")
	}

	rolesInterface, ok := realmAccess["roles"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("roles not found in realm_access")
	}

	// Extract all project UUIDs from roles
	projectIDs := make(map[string]bool)
	for _, roleInterface := range rolesInterface {
		role, ok := roleInterface.(string)
		if !ok {
			continue
		}

		// Roles with project context follow pattern: {projectUUID}_{roleName}
		if strings.Contains(role, roleProjectIDSeparator) {
			parts := strings.Split(role, roleProjectIDSeparator)
			if len(parts) > 0 {
				potentialUUID := parts[0]
				if isValidUUID(potentialUUID) {
					projectIDs[potentialUUID] = true
				}
			}
		}
	}

	if len(projectIDs) == 0 {
		return nil, fmt.Errorf("no project IDs found in JWT roles")
	}

	return projectIDs, nil
}

// ValidateProjectAccess checks if the ProjectID is accessible under the given user's JWT token without signature verification.
// It extracts all project UUIDs from the JWT roles and verifies that projectID is among them.
// This prevents users from accessing projects they don't have permissions for.
// This function assumes the JWT has already been validated by upstream authentication middleware.
// Use ValidateProjectAccessWithVerification if you need signature verification.
func ValidateProjectAccess(authHeader string, projectID string) error {
	return validateProjectAccess(authHeader, projectID, false)
}

// ValidateProjectAccessWithVerification checks if the ProjectID is accessible under the given user's JWT token
// and verifies the JWT signature using ORCH_JWT_SIGNING_KEY.
// Use this when the JWT has NOT been validated by upstream middleware.
func ValidateProjectAccessWithVerification(authHeader string, projectID string) error {
	return validateProjectAccess(authHeader, projectID, true)
}

// validateProjectAccess is the internal implementation that handles both verified and unverified validation.
func validateProjectAccess(authHeader string, projectID string, verifySignature bool) error {
	if authHeader == "" {
		return fmt.Errorf("missing authorization header")
	}

	if projectID == "" {
		return fmt.Errorf("missing project ID")
	}

	// Extract all project IDs from JWT token
	projectIDs, err := extractAllProjectIDsFromJWT(authHeader, verifySignature)
	if err != nil {
		return fmt.Errorf("failed to extract projects from JWT: %w", err)
	}

	// Check if the requested project ID is in the user's accessible projects
	if _, exists := projectIDs[projectID]; !exists {
		return fmt.Errorf("user does not have access to project: %s", projectID)
	}

	return nil
}

func isValidUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}
