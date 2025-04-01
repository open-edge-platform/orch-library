// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"gotest.tools/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const dexWellKnownOpenIDConfig = `{
  "issuer": "http://dex:32000",
  "authorization_endpoint": "http://dex:32000/auth",
  "token_endpoint": "http://dex:32000/token",
  "jwks_uri": "http://dex:32000/keys",
  "userinfo_endpoint": "http://dex:32000/userinfo",
  "response_types_supported": [
    "code"
  ],
  "subject_types_supported": [
    "public"
  ],
  "id_token_signing_alg_values_supported": [
    "RS256"
  ],
  "scopes_supported": [
    "openid",
    "email",
    "groups",
    "profile",
    "offline_access"
  ],
  "token_endpoint_auth_methods_supported": [
    "client_secret_basic"
  ],
  "claims_supported": [
    "aud",
    "email",
    "email_verified",
    "exp",
    "iat",
    "iss",
    "locale",
    "name",
    "sub"
  ]
}`

// Available from a running Dex instance at http://dex:32000/keys
const dexkeys = `{
  "keys": [
    {
      "use": "sig",
      "kty": "RSA",
      "kid": "df362dc374c5bf1a2f1039ed8d6629c90b93a04e",
      "alg": "RS256",
      "n": "uWnUNbMHMdi6UH3AKEfslIBoOpg7fFkMG108BhjnY1YVXBPAZ_2QbdY7cA04mPNU4GJgu0FVQeZKsDB9biCTv7QKwDnBqGldBku27V5R9AE3icrDspHjfUSVu-hFSY_hWz_UfHbnfUHknpGp35MO7J2x0A7CvvBcT1LrQBihpY09JBrPXIPfTtygqhsHJ_aejbak_R6HRVvf6iyrHSuanTSKRpXpiFypfP3hsWRy36sO2cEnBYV4t-vNG7Xi8hTPK-hul5DDCpeUu2QpPma1GnTEszOHBdFWNHl5rdrIgkP7RnKmugnbxRAaQChhJrE-SWSR0FrHZsioQx94FqaCAw",
      "e": "AQAB"
    },
    {
      "use": "sig",
      "kty": "RSA",
      "kid": "896728487539b52bcda1de629c6fd8263193db85",
      "alg": "RS256",
      "n": "qxcVBYBG42H5m2T4WdKloq7Cyva1YOI4FLWgHUlnQ17yYp7tXVTRtFKGeKC4uprhM2SA_KAbXtvXEq2cDoXpBDk1LwiKU0IUvVoi6kYQiJDYvXalP9H57OmiTXMwpB7ZPnEupBJqBUy8Nw34b7EOCKo3BLkcV_aczl1jkpROi5U0tGmuaBVjpigQ4Q9CdnHeKGRExPjStd-XgTJkXHxfnSt5EW5DWc3HPAR911_DUoXTElaYrZDgW8XDcYUgCuifHikgwahUmKoVIJMCqvvfIZPEYCl_Kp4VUMI46L6K_aRadTs1-gaodP0O-rJWu1Of7YuLwAdl0Hrw4eRZ46crew",
      "e": "AQAB"
    },
    {
      "use": "sig",
      "kty": "RSA",
      "kid": "ad458979046bfdcdad0db2250f10f526366d461d",
      "alg": "RS256",
      "n": "3ua3Mea1y9Q67oSS2cVBAfpvlkY0E6_1hyoOLZalVX2Wwnm113l60Blbk8UK3zKlBIfZjF4mwUrz8aj2nrsLOz7Wle-X2Gu9YsjZkdX0aoZsLplGXPnVyt10hkD6LPr-yyvqcJGNBnDot4UjNRoHfU0rsTc0BIR2XmVd9YCmUx7Rfi2GLrDStEX_RLU8aVYXeKlazfbqPhsm43muugzSCNaBF28_wtDjjPT8WTmOv-EyE0lZEv2jJuwFN6mtJzN0pXy2cj9oS-nhv0_pE635qUvrX0h3-8Uq9IwQQFYGaCQcfLRgyYXO3oSgqyw_Tt8anaMpjBOjmPgNSDOFPd47Iw",
      "e": "AQAB"
    }
  ]
}`

const sharedSecretKey = `ZXhhbXBsZS1hcHAtc2VjcmV0` //nolint: gosec

// generated from running onos-gui against the Dex IDP - expires 12 Jul 2030
const sampleTokenOnosGuiAndDex = `eyJhbGciOiJSUzI1NiIsImtpZCI6ImRmMzYyZGMzNzRjNWJmMWEyZjEwMzllZDhkNjYyOWM5MGI5M2EwNGUifQ.eyJpc3MiOiJodHRwOi8vZGV4OjMyMDAwIiwic3ViIjoiQ2lRd09HRTROamcwWWkxa1lqZzRMVFJpTnpNdE9UQmhPUzB6WTJReE5qWXhaalUwTmpnU0JXeHZZMkZzIiwiYXVkIjoib25vcy1ndWkiLCJleHAiOjE5MDk5Mzg0NDcsImlhdCI6MTU5NDU3ODQ0Nywibm9uY2UiOiJjelJ2TmtoTFVYbFhVSE5EYWpsMmNHSlFjMVZOUld0T2NERmtMbkJ0ZHkxbFJTMVlkM0YzVUVGNE5IZFEiLCJhdF9oYXNoIjoiWXcxQ1M1UmdvYVRVWUNUZWtxUkxBZyIsImVtYWlsIjoic2VhbkBvcGVubmV0d29ya2luZy5vcmciLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6InNlYW4ifQ.OMxzPd9CD7CoZVPRYSXJdLy8QCQjkIyjD7xFKBbDDsOCvIK9MkNHBTgQfrc3gmv8DZaAyzj9abLRS1TPWBwfQl5QJW9unXbjtk4KukXxv0CbKN3NyRV5Nm5YlGm66DIlPMj8udlqyau_xUJXrVC4T13sPo1CVpnAHut6Rr9zwPj3pVLwPXO2dkqHH4c1YM9Lyg8fpMv5eGd3iN6xcRMCcxUFkqffvwS1mCW6BUoBvnbMMZEaNokAC6HcWD8EB_m-Z7nt_xP_C_mnnZGFJNGDU0fRUjdGRxQ6AHYVg1zCS_B8P1IxkIe6tnBRi8309s99Q4MhlDWSoju_fe8pU7iLwQ` //nolint: gosec
const sampleTokenHS256Signature = `eyJhbGciOiJIUzI1NiIsImtpZCI6ImRmMzYyZGMzNzRjNWJmMWEyZjEwMzllZDhkNjYyOWM5MGI5M2EwNGUifQ.eyJpc3MiOiJodHRwOi8vZGV4OjMyMDAwIiwic3ViIjoiQ2lRd09HRTROamcwWWkxa1lqZzRMVFJpTnpNdE9UQmhPUzB6WTJReE5qWXhaalUwTmpnU0JXeHZZMkZzIiwiYXVkIjoib25vcy1ndWkiLCJleHAiOjE5MDk5Mzg0NDcsImlhdCI6MTU5NDU3ODQ0Nywibm9uY2UiOiJjelJ2TmtoTFVYbFhVSE5EYWpsMmNHSlFjMVZOUld0T2NERmtMbkJ0ZHkxbFJTMVlkM0YzVUVGNE5IZFEiLCJhdF9oYXNoIjoiWXcxQ1M1UmdvYVRVWUNUZWtxUkxBZyIsImVtYWlsIjoic2VhbkBvcGVubmV0d29ya2luZy5vcmciLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6InNlYW4ifQ.vAx1WCxsJNEvpfMvTU0IqnWgen1xmS3lOjkorZL5uWU`                                                                                                                                                                                                                                                                                                           //nolint: gosec

func TestJwtAuthenticator_parseToken(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case "/" + OpenidConfiguration:
			fmt.Fprintln(w, strings.ReplaceAll(dexWellKnownOpenIDConfig, "dex:32000", r.Host))
		case "/keys":
			fmt.Fprintln(w, dexkeys)
		default:
			t.Fatalf("Unexpected URL %s", r.URL.String())
		}
	}))
	defer ts.Close()

	_ = os.Setenv(OIDCServerURL, ts.URL)

	authenticator := new(JwtAuthenticator)

	claims, err := authenticator.ParseAndValidate(sampleTokenOnosGuiAndDex)
	assert.NilError(t, err, "unexpected error parsing token")
	assert.Assert(t, claims != nil)

	issuedAt, err := claims.GetIssuedAt()
	assert.NilError(t, err)
	assert.Equal(t, int64(1594578447), issuedAt.Unix(), "error verifying issuedat time")
	expiresAt, err := claims.GetExpirationTime()
	assert.NilError(t, err)
	assert.Equal(t, int64(1909938447), expiresAt.Unix(), "error verifying expiry time")

	issuer, err := claims.GetIssuer()
	assert.NilError(t, err)
	assert.Equal(t, `http://dex:32000`, issuer, "error verifying issuer")

	audience, err := claims.GetAudience()
	assert.NilError(t, err)
	audienceJSON, err := audience.MarshalJSON()
	assert.NilError(t, err)
	assert.Equal(t, `["onos-gui"]`, string(audienceJSON), "error verifying audience")

	claimsMap, isMap := claims.(jwt.MapClaims)
	assert.Assert(t, isMap)
	name, ok := claimsMap["name"]
	assert.Assert(t, ok, "error extracting name")
	assert.Equal(t, "sean", name, "error unexpected name", name)
	email, ok := claimsMap["email"]
	assert.Assert(t, ok, "error extracting email")
	assert.Equal(t, "sean@opennetworking.org", email, "error unexpected email", email)

}

func TestJwtAuthenticator_HSAlgorithm(t *testing.T) {
	authenticator := new(JwtAuthenticator)

	err := os.Setenv("SHARED_SECRET_KEY", sharedSecretKey)
	assert.NilError(t, err, "shared secret key is not set")

	claims, err := authenticator.ParseAndValidate(sampleTokenHS256Signature)
	assert.NilError(t, err, "unexpected error parsing token")

	issuedAt, err := claims.GetIssuedAt()
	assert.NilError(t, err)
	assert.Equal(t, int64(1594578447), issuedAt.Unix(), "error verifying issuedat time")
	expiresAt, err := claims.GetExpirationTime()
	assert.NilError(t, err)
	assert.Equal(t, int64(1909938447), expiresAt.Unix(), "error verifying expiry time")

	issuer, err := claims.GetIssuer()
	assert.NilError(t, err)
	assert.Equal(t, `http://dex:32000`, issuer, "error verifying issuer")

	audience, err := claims.GetAudience()
	assert.NilError(t, err)
	audienceJSON, err := audience.MarshalJSON()
	assert.NilError(t, err)
	assert.Equal(t, `["onos-gui"]`, string(audienceJSON), "error verifying audience")

	claimsMap, isMap := claims.(jwt.MapClaims)
	assert.Assert(t, isMap)
	name, ok := claimsMap["name"]
	assert.Assert(t, ok, "error extracting name")
	assert.Equal(t, "sean", name, "error unexpected name", name)
	email, ok := claimsMap["email"]
	assert.Assert(t, ok, "error extracting email")
	assert.Equal(t, "sean@opennetworking.org", email, "error unexpected email", email)

}
