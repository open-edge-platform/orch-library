// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

type M2MTestSuite struct {
	suite.Suite
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *M2MTestSuite) SetupSuite() {
}

func (s *M2MTestSuite) TearDownSuite() {
}

func (s *M2MTestSuite) SetupTest() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 3*time.Minute)
}

func (s *M2MTestSuite) TearDownTest() {
}

type TestHTTPServer struct {
	K8SLoginReadHandler  func(w http.ResponseWriter)
	SecretHandler        func(w http.ResponseWriter, r *http.Request)
	RevokeHandler        func(w http.ResponseWriter)
	KeycloakTokenHandler func(w http.ResponseWriter, r *http.Request)
	Server               *httptest.Server
}

func (t *TestHTTPServer) WithK8SLoginReadHandler(K8SLoginReadHANDLER func(w http.ResponseWriter)) *TestHTTPServer {
	t.K8SLoginReadHandler = K8SLoginReadHANDLER
	return t
}

func (t *TestHTTPServer) WithSecretHandler(SecretHandler func(w http.ResponseWriter, r *http.Request)) *TestHTTPServer {
	t.SecretHandler = SecretHandler
	return t
}

func (t *TestHTTPServer) WithKeycloakTokenHandler(KeycloakTokenHandler func(w http.ResponseWriter, r *http.Request)) *TestHTTPServer {
	t.KeycloakTokenHandler = KeycloakTokenHandler
	return t
}

func (t *TestHTTPServer) WithRevokeHandler(RevokeHandler func(w http.ResponseWriter)) *TestHTTPServer {
	t.RevokeHandler = RevokeHandler
	return t
}

func (s *M2MTestSuite) NewTestHTTPServer() *TestHTTPServer {
	return &TestHTTPServer{
		K8SLoginReadHandler:  s.handleK8SLogin,
		SecretHandler:        s.handleSecret,
		RevokeHandler:        s.handleRevoke,
		KeycloakTokenHandler: s.handleKeycloakToken,
	}
}

var (
	VaultServer    string
	KeycloakServer string
)

func (t *TestHTTPServer) Start() *TestHTTPServer {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case vaultK8SLoginURL:
			t.K8SLoginReadHandler(w)
		case vaultSecretBaseURL + `catalog-bootstrap-m2m-client-secret`:
			t.SecretHandler(w, r)
		case vaultRevokeSelfURL:
			t.RevokeHandler(w)
		case keycloakTokenURL:
			t.KeycloakTokenHandler(w, r)
		}
	}))
	secrets[vaultSecretBaseURL+`catalog-bootstrap-m2m-client-secret`] = `{"data":{"data":{"value":"` + `secret` + `"}}}`
	t.Server = server
	VaultServer = server.URL
	KeycloakServer = server.URL
	K8STokenFile = `testdata/k8stoken` // #nosec
	return t
}

func (t *TestHTTPServer) Stop() {
	t.Server.Close()
}

func (s *M2MTestSuite) handleK8SLogin(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	var loginResp struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
		Errors []string `json:"errors"`
	}
	loginResp.Auth.ClientToken = "token"
	js, err := json.Marshal(loginResp)
	s.NoError(err)
	_, _ = w.Write(js)
}

func (s *M2MTestSuite) HandleK8SLoginBadJSON(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	var loginResp struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
		Errors []string `json:"errors"`
	}
	loginResp.Auth.ClientToken = "token"
	_, _ = w.Write([]byte("This is not the JSON you are looking for"))
}

func (s *M2MTestSuite) handleRevoke(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *M2MTestSuite) HandleRevokeHTTPError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

var secrets = map[string]string{}

func (s *M2MTestSuite) handleSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusOK)
		secretData := map[string]interface{}{}
		rawData, err := io.ReadAll(r.Body)
		s.NoError(err)
		err = json.Unmarshal(rawData, &secretData)
		s.NoError(err)
		data := secretData["data"].(map[string]interface{})
		s.NotNil(data)
		secret := data["value"].(string)
		secretJSON := `{"data":{"data":{"value":"` + secret + `"}}}`
		secrets[r.URL.Path] = secretJSON
	} else if r.Method == http.MethodGet {
		secretJSON, ok := secrets[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(secretJSON))
		}
	} else if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusOK)
		delete(secrets, r.URL.Path)
	}
}

func (s *M2MTestSuite) handleKeycloakToken(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var tokenResp struct {
			AccessToken string `json:"access_token"`
		}
		tokenResp.AccessToken = "token"
		w.WriteHeader(http.StatusOK)
		b, err := json.Marshal(tokenResp)
		s.NoError(err)
		count, err := w.Write(b)
		s.NoError(err)
		s.Len(b, count)
	}
}

func (s *M2MTestSuite) HandleSecretBadHTTPStatus(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func (s *M2MTestSuite) TestGetM2MToken() {
	server := s.NewTestHTTPServer().Start()
	defer server.Stop()

	v, err := NewVaultAuth(KeycloakServer, VaultServer, "test-svc")
	s.NoError(err)
	s.NotNil(v)
	s.NoError(os.Setenv("USE_M2M_TOKEN", "true"))
	token, err := v.GetM2MToken(context.Background())
	s.NoError(err)
	s.NotNil(token)
}

func (s *M2MTestSuite) TestCreateClientSecretNoKeycloak() {
	server := s.NewTestHTTPServer().Start()
	defer server.Stop()

	v, err := NewVaultAuth(KeycloakServer, VaultServer, "test-svc")
	s.NoError(err)
	s.NotNil(v)
	s.NoError(os.Setenv("USE_M2M_TOKEN", "true"))
	secret, err := v.CreateClientSecret(context.Background(), "u", "p")
	s.Error(err)
	s.Equal("", secret)
}

func TestM2M(t *testing.T) {
	suite.Run(t, &M2MTestSuite{})
}
