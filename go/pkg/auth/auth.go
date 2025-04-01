// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package auth

import "github.com/golang-jwt/jwt/v5"

// Authenticator an authenticator interface to implement different authentication methods
type Authenticator interface {
	// Authenticate authenticates a given string token
	Authenticate(string) (jwt.MapClaims, error)
}
