// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0
package projectcontext

import "regexp"


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


// ValidateTenantAccess checks if the given tenant has access to the specified project
// pkg/auth/project.go


// InjectActiveProjectID injects the active project UUID into the gRPC metadata