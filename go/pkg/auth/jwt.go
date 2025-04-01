// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/open-edge-platform/orch-library/go/dazl"

	"io"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc/codes"
)

var log = dazl.GetLogger()

var publicKeys map[string][]byte

const (
	// SharedSecretKey shared secret key for signing a token
	SharedSecretKey = "SHARED_SECRET_KEY"
	// OIDCServerURL - will be accessed as Environment variable
	OIDCServerURL = "OIDC_SERVER_URL"

	// OIDCTlsInsecureSkipVerify - will be accessed as Environment variable
	OIDCTlsInsecureSkipVerify = "OIDC_TLS_INSECURE_SKIP_VERIFY"

	// OpenidConfiguration is the discovery point on the OIDC server
	OpenidConfiguration = ".well-known/openid-configuration"
	// HS prefix for HS family algorithms
	HS = "HS"
	// RS prefix for RS family algorithms
	RS = "RS"
	// PS prefix for PS family algorithms
	PS = "PS"
)

type providerJSON struct {
	Issuer        string   `json:"issuer"`
	AuthURL       string   `json:"authorization_endpoint"`
	TokenURL      string   `json:"token_endpoint"`
	DeviceAuthURL string   `json:"device_authorization_endpoint"`
	JWKSURL       string   `json:"jwks_uri"`
	UserInfoURL   string   `json:"userinfo_endpoint"`
	Algorithms    []string `json:"id_token_signing_alg_values_supported"`
}

func init() {
	publicKeys = make(map[string][]byte)
	if err := refreshJwksKeys(); err != nil {
		log.Debugf("unable to refresh JWKS keys on init %s", err)
	}
}

// JwtAuthenticator jwt authenticator
type JwtAuthenticator struct {
}

// ParseToken parse token and Ensure that the JWT conforms to the structure of a JWT.
func (j *JwtAuthenticator) parseToken(tokenString string) (*jwt.Token, jwt.Claims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// HS256, HS384, or HS512
		if strings.HasPrefix(token.Method.Alg(), HS) {
			key := os.Getenv(SharedSecretKey)
			return []byte(key), nil
			// ES256, ES384, ES512, PS256, PS384, PS512, RS256, RS384, RS512
		} else if strings.HasPrefix(token.Method.Alg(), RS) || strings.HasPrefix(token.Method.Alg(), PS) {
			keyID, ok := token.Header["kid"]
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "token header not found 'kid' (key ID)")
			}
			keyIDStr := keyID.(string)
			publicKey, ok := publicKeys[keyIDStr]
			if !ok {
				// Keys may have been refreshed on the server
				// Fetch them again and try once more before failing
				if err := refreshJwksKeys(); err != nil {
					return nil, status.Errorf(codes.Unauthenticated, "unable to refresh keys from ID provider %s", err)
				}
				// try again after refresh
				if publicKey, ok = publicKeys[keyIDStr]; !ok {
					return nil, status.Errorf(codes.Unauthenticated, "token has obsolete key ID %s", keyID)
				}
			}
			rsaPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, err.Error())
			}
			return rsaPublicKey, nil
		}
		return nil, status.Errorf(codes.Unauthenticated, "unknown signing algorithm: %s", token.Method.Alg())
	})

	return token, claims, err

}

// ParseAndValidate parse a jwt string token and validate it
func (j *JwtAuthenticator) ParseAndValidate(tokenString string) (jwt.Claims, error) {
	token, claims, err := j.parseToken(tokenString)
	if err != nil {
		log.Warnf("cannot parse token. %s", err.Error())
		return nil, err
	}

	// Check the token is valid
	if !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "token is not valid %v", token)
	}

	return claims, nil
}

// Connect back to the OpenIDConnect server to retrieve the keys
// They are rotated every 6 hours by default - we keep the keys in a cache
// It's a 2 step process
// 1) connect to $OIDCServerURL/.well-known/openid-configuration and retrieve the JSON payload
// 2) lookup the "keys" parameter and get keys from $OIDCServerURL/keys
// The keys are in a public key format and are converted to RSA Public Keys
func refreshJwksKeys() error {
	oidcURL, present := os.LookupEnv(OIDCServerURL)
	if !present {
		return fmt.Errorf("environmental variable OIDC_SERVER_URL is not set " +
			"Can't reach the OIDC server to refresh JWKS")
	}

	oidcClient := new(http.Client)

	oidcTLSInsecureSkipVerify := os.Getenv(OIDCTlsInsecureSkipVerify)
	if strings.ToLower(oidcTLSInsecureSkipVerify) == "true" {
		oidcClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS12,
			},
		}
	}

	resOpenIDConfig, err := oidcClient.Get(fmt.Sprintf("%s/%s", oidcURL, OpenidConfiguration))
	if err != nil {
		return fmt.Errorf("error obtaining information from OIDC well-known URL: %v", err)
	}
	if resOpenIDConfig.Body != nil {
		defer resOpenIDConfig.Body.Close()
	}
	openIDConfigBody, readErr := io.ReadAll(resOpenIDConfig.Body)
	if readErr != nil {
		return fmt.Errorf("error reading Body of the OIDC configuration: %v", readErr)
	}
	var openIDprovider providerJSON
	jsonErr := json.Unmarshal(openIDConfigBody, &openIDprovider)
	if jsonErr != nil {
		return fmt.Errorf("error unmarshalling OIDC configuration: %v", jsonErr)
	}
	resOpenIDKeys, err := oidcClient.Get(openIDprovider.JWKSURL)
	if err != nil {
		return fmt.Errorf("error retrieving JWKS from the OIDC provider: %v", err)
	}
	if resOpenIDKeys.Body != nil {
		defer resOpenIDKeys.Body.Close()
	}
	bodyOpenIDKeys, readErr := io.ReadAll(resOpenIDKeys.Body)
	if readErr != nil {
		return fmt.Errorf("error reading keys from the Body of the response: %v", readErr)
	}
	var jsonWebKeySet jose.JSONWebKeySet
	if err := json.Unmarshal(bodyOpenIDKeys, &jsonWebKeySet); err != nil {
		return fmt.Errorf("error unmarshalling JWKS to JSON: %v", err)
	}

	// Clear out old keys
	for k := range publicKeys {
		delete(publicKeys, k)
	}

	for _, key := range jsonWebKeySet.Keys {
		data, err := x509.MarshalPKIXPublicKey(key.Key)
		if err != nil {
			return err
		}
		block := pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: data,
		}
		pemBytes := pem.EncodeToMemory(&block)
		publicKeys[key.KeyID] = pemBytes
	}
	log.Infof("Refreshed JWKS keys from %s", openIDprovider.JWKSURL)

	return nil
}
