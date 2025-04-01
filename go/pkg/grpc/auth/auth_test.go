// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/open-edge-platform/orch-library/go/pkg/auth"
	"google.golang.org/grpc/metadata"
	"gotest.tools/assert"
	"os"
	"strings"
	"testing"
	"time"
)

type RealmAccess struct {
	Roles []string
}

type Account struct {
	Roles []string `json:"roles"`
}

type ResourceAccess struct {
	Account Account `json:"account"`
}
type TestCustomClaims struct {
	jwt.RegisteredClaims
	Name              string         `json:"name"`
	Email             string         `json:"email"`
	EmailVerified     bool           `json:"email_verified"`
	PreferredUsername string         `json:"preferred_username"`
	Groups            []string       `json:"groups"`
	Roles             []string       `json:"roles"`
	Foo               int            `json:"foo"`
	Foo32             int32          `json:"foo32"`
	RealmAccess       RealmAccess    `json:"realm-access"`
	ResourceAccess    ResourceAccess `json:"resource-access"`
}

func createCustomClaims() TestCustomClaims {
	now := time.Now()
	return TestCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://dex:32000",
			Subject:   "Test_AuthenticationInterceptor",
			Audience:  []string{"testaudience"},
			ExpiresAt: &jwt.NumericDate{Time: now.Add(24 * time.Hour)},
			NotBefore: &jwt.NumericDate{Time: now},
			IssuedAt:  &jwt.NumericDate{Time: now},
			ID:        "",
		},
		Name:              "testname",
		Email:             "test1@opennetworking.org",
		EmailVerified:     true,
		PreferredUsername: "a user Name",
		Groups:            []string{"testGroup1", "testGroup2"},
		Roles:             []string{"testRole1", "testRole2"},
		Foo:               21,
		Foo32:             22,
		RealmAccess: RealmAccess{
			Roles: []string{
				"testRole1",
				"testRole2",
			},
		},
		ResourceAccess: ResourceAccess{
			Account: Account{
				Roles: []string{
					"testRole1",
					"testRole2",
				},
			},
		},
	}
}

func (c TestCustomClaims) Validate() error {
	if c.Name == "" || c.Email == "" {
		return fmt.Errorf("name or email cannot be empty")
	}
	return nil
}

func Test_AuthenticationInterceptor(t *testing.T) {
	claims := createCustomClaims()
	assert.NilError(t, claims.Validate())

	signingKey := "testkey"
	err := os.Setenv(auth.SharedSecretKey, signingKey)
	assert.NilError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(signingKey))
	assert.NilError(t, err)

	mdIn := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", s))
	ctx := metadata.NewIncomingContext(context.Background(), mdIn)
	intercepted, err := AuthenticationInterceptor(ctx)
	assert.NilError(t, err)
	// We're not using github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata in the test
	// Just using the plain google.golang.org/grpc/metadata
	// This is because the former only returns the first entry of a value when using Get()
	md, ok := metadata.FromIncomingContext(intercepted)
	assert.Assert(t, ok)
	assert.Assert(t, md != nil, "expected a value for Metadata")
	assert.DeepEqual(t, []string{"testname"}, md.Get("Name"))
	assert.DeepEqual(t, []string{"test1@opennetworking.org"}, md.Get("Email"))
	assert.DeepEqual(t, []string{"testGroup1", "testGroup2"}, md.Get("Groups"))
	assert.DeepEqual(t, []string{"testRole1", "testRole2"}, md.Get("Roles"))
	assert.DeepEqual(t, []string{"a user Name"}, md.Get("preferred_username"))
	assert.Assert(t, strings.HasPrefix(md.Get("authorization")[0], ContextMetadataBearerKeyLower))
	assert.DeepEqual(t, []string{"testRole1", "testRole2"}, md.Get("realm-access/roles"))
	assert.DeepEqual(t, []string{"testRole1", "testRole2"}, md.Get("resource-access/account/roles"))

	mdIn2 := metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", s))
	ctx2 := metadata.NewIncomingContext(context.Background(), mdIn2)
	intercepted2, err2 := AuthenticationInterceptor(ctx2)
	assert.NilError(t, err2)
	// We're not using github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata in the test
	// Just using the plain google.golang.org/grpc/metadata
	// This is because the former only returns the first entry of a value when using Get()
	md2, ok2 := metadata.FromIncomingContext(intercepted2)
	assert.Assert(t, ok2)
	assert.Assert(t, md2 != nil, "expected a value for Metadata")
	assert.DeepEqual(t, []string{"testname"}, md2.Get("Name"))
	assert.DeepEqual(t, []string{"test1@opennetworking.org"}, md2.Get("Email"))
	assert.DeepEqual(t, []string{"testGroup1", "testGroup2"}, md2.Get("Groups"))
	assert.DeepEqual(t, []string{"testRole1", "testRole2"}, md2.Get("Roles"))
	assert.DeepEqual(t, []string{"a user Name"}, md2.Get("preferred_username"))
	assert.Assert(t, strings.HasPrefix(md2.Get("authorization")[0], ContextMetadataBearerKeyCamel))
	assert.DeepEqual(t, []string{"testRole1", "testRole2"}, md2.Get("realm-access/roles"))
	assert.DeepEqual(t, []string{"testRole1", "testRole2"}, md2.Get("resource-access/account/roles"))
}

func Test_AuthenticationInterceptor_InvalidExpiry(t *testing.T) {
	now := time.Now()

	claims := TestCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://dex:32000",
			Audience:  []string{"testaudience"},
			ExpiresAt: &jwt.NumericDate{Time: now.Add(-24 * time.Hour)}, // Before issued at
			NotBefore: &jwt.NumericDate{Time: now},
			IssuedAt:  &jwt.NumericDate{Time: now},
		},
		Name:  "testname",
		Email: "test1@opennetworking.org",
	}
	assert.NilError(t, claims.Validate())

	signingKey := "testkey"
	err := os.Setenv(auth.SharedSecretKey, signingKey)
	assert.NilError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(signingKey))
	assert.NilError(t, err)

	mdIn := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", s))
	ctx := metadata.NewIncomingContext(context.Background(), mdIn)
	_, err = AuthenticationInterceptor(ctx)
	assert.ErrorContains(t, err, "token has invalid claims: token is expired")
}

func Test_AuthenticationInterceptor_NoAuth_NotAllowed(t *testing.T) {
	oldValue := os.Getenv(allowMissingAuthClients)
	assert.NilError(t, os.Setenv(allowMissingAuthClients, "some-other-client"))
	defer func() {
		assert.NilError(t, os.Setenv(allowMissingAuthClients, oldValue))
	}()

	signingKey := "testkey"
	err := os.Setenv(auth.SharedSecretKey, signingKey)
	assert.NilError(t, err)

	mdIn := metadata.Pairs(ContextMetadataClientKeyLower, "test-client")
	ctx := metadata.NewIncomingContext(context.Background(), mdIn)
	_, err = AuthenticationInterceptor(ctx)
	assert.ErrorContains(t, err, "Request unauthenticated with bearer")
}

func Test_AuthenticationInterceptor_NoAuth_Allowed(t *testing.T) {
	oldValue := os.Getenv(allowMissingAuthClients)
	assert.NilError(t, os.Setenv(allowMissingAuthClients, "test-client,some-other-client"))
	defer func() {
		assert.NilError(t, os.Setenv(allowMissingAuthClients, oldValue))
	}()

	signingKey := "testkey"
	err := os.Setenv(auth.SharedSecretKey, signingKey)
	assert.NilError(t, err)

	mdIn := metadata.Pairs(ContextMetadataClientKeyCamel, "test-client")
	ctx := metadata.NewIncomingContext(context.Background(), mdIn)
	intercepted, err := AuthenticationInterceptor(ctx)
	assert.NilError(t, err)
	md, ok := metadata.FromIncomingContext(intercepted)
	assert.Assert(t, ok)
	assert.DeepEqual(t, []string(nil), md.Get("iat"))
}

func Test_HandleClaims(t *testing.T) {
	// setting context
	ctx := context.Background()
	niceMd := metautils.ExtractIncoming(ctx)

	// creating claims
	claims := createCustomClaims()
	assert.NilError(t, claims.Validate())

	// marshalling claims to JSON
	jsonBytes, err := json.Marshal(claims)
	assert.NilError(t, err)

	// unmarshalling claims to claimsMap, which is type-agnostic
	var claimsMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &claimsMap)
	assert.NilError(t, err)

	// mimic piece of code from Interceptor
	for k, v := range claimsMap {
		err = HandleClaim(&niceMd, []string{k}, v)
		assert.NilError(t, err)
	}

	// extracting claims added to the context and verifying that all are present
	md, ok := metadata.FromIncomingContext(niceMd.ToIncoming(ctx))
	assert.Assert(t, ok)
	assert.Assert(t, md != nil, "expected a value for Metadata")
	assert.DeepEqual(t, []string{"testname"}, md.Get("Name"))
	assert.DeepEqual(t, []string{"test1@opennetworking.org"}, md.Get("Email"))
	assert.DeepEqual(t, []string{"testGroup1", "testGroup2"}, md.Get("Groups"))
	assert.DeepEqual(t, []string{"testRole1", "testRole2"}, md.Get("Roles"))
	assert.DeepEqual(t, []string{"a user Name"}, md.Get("preferred_username"))
	assert.DeepEqual(t, []string{"testRole1", "testRole2"}, md.Get("realm-access/roles"))
	assert.DeepEqual(t, []string{"testRole1", "testRole2"}, md.Get("resource-access/account/roles"))
}
