// SPDX-FileCopyrightText: (C) 2023 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	keycloakTokenURL       = "/realms/master/protocol/openid-connect/token"
	keycloakAdminClient    = "system-client"
	keycloakUserClientName = "edge-manager-m2m-client"

	vaultK8STokenFile  = `/var/run/secrets/kubernetes.io/serviceaccount/token` // #nosec G101
	vaultK8SLoginURL   = `/v1/auth/kubernetes/login`
	vaultSecretBaseURL = `/v1/secret/data/`           // #nosec
	vaultRevokeSelfURL = `/v1/auth/token/revoke-self` // #nosec
	m2mVaultSecretPath = "catalog-bootstrap-m2m-client-secret"
)

var (
	K8STokenFile = vaultK8STokenFile
)

// Can be overridden for read unit testing
func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

var readAllFactory = readAll

//go:generate mockery --name VaultAuth --filename vault_mock.go --structname VaultAuth
type VaultAuth interface {
	GetVaultToken(ctx context.Context) (string, error)
	GetM2MToken(ctx context.Context) (string, error)
	CreateClientSecret(ctx context.Context, username string, password string) (string, error)
	Logout(ctx context.Context) error
}

type vaultAuth struct {
	keycloakServer string
	vaultServer    string
	serviceAccount string
	httpClient     *http.Client
	vaultToken     string
	mu             sync.Mutex
}

func NewVaultAuth(keycloakServer string, vaultServer string, serviceAccount string) (VaultAuth, error) {
	client, err := getHTTPClient()
	if err != nil {
		return nil, err
	}
	auth := &vaultAuth{
		httpClient:     client,
		keycloakServer: keycloakServer,
		vaultServer:    vaultServer,
		serviceAccount: serviceAccount,
	}
	return auth, nil
}

func getHTTPClient() (*http.Client, error) {
	return &http.Client{
		Timeout: 10 * time.Second,
	}, nil
}

func (v *vaultAuth) httpsVaultURL(path string) string {
	return v.vaultServer + path
}

// GetVaultToken reads and returns the secret token to access the vault
func (v *vaultAuth) GetVaultToken(ctx context.Context) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	tokenData, err := os.ReadFile(K8STokenFile)
	if err != nil {
		return "", err
	}
	loginReq := struct {
		JWT  string `json:"jwt"`
		Role string `json:"role"`
	}{
		JWT:  string(tokenData),
		Role: v.serviceAccount,
	}
	body, _ := json.Marshal(loginReq)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		v.httpsVaultURL(vaultK8SLoginURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("vault login request failed %d", resp.StatusCode)
	}

	var loginResp struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
		Errors []string `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)

	if err != nil {
		return "", err
	}
	if loginResp.Auth.ClientToken == "" {
		return "", fmt.Errorf("unable to get client token: %v", loginResp)
	}
	v.vaultToken = loginResp.Auth.ClientToken
	return v.vaultToken, nil
}

type ClientSecretData struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// getClientSecretFromVault reads and returns the client secret from the vault
func (v *vaultAuth) getClientSecretFromVault(ctx context.Context, httpClient *http.Client, vaultClientToken string) (string, error) {
	vaultURL := v.httpsVaultURL(vaultSecretBaseURL + m2mVaultSecretPath)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		vaultURL,
		nil,
	)
	if err != nil {
		return "", err
	}

	req.Header.Add("X-Vault-Token", vaultClientToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	getRawData, err := readAllFactory(resp.Body)
	if err != nil {
		return "", err
	}

	var secret vault.Secret
	err = json.Unmarshal(getRawData, &secret)
	if err != nil {
		return "", err
	}

	type VaultPathGetResponse struct {
		Data struct {
			Data struct {
				ClientID     string `json:"client_id"`
				ClientSecret string `json:"client_secret"`
			} `json:"data"`
		} `json:"data"`
	}
	var getResult VaultPathGetResponse
	err = json.Unmarshal(getRawData, &getResult)
	if err != nil {
		return "", err
	}

	return getResult.Data.Data.ClientSecret, nil
}

func (v *vaultAuth) getAdminTokenFromKeycloak(ctx context.Context, httpClient *http.Client, adminUsername string, adminPassword string) (string, error) {
	keycloakURL := v.keycloakServer + keycloakTokenURL
	vals := url.Values{}
	vals.Add("grant_type", "password")
	vals.Add("client_id", keycloakAdminClient)
	vals.Add("username", adminUsername)
	vals.Add("password", adminPassword)
	vals.Add("scope", "openid profile email groups")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		keycloakURL, strings.NewReader(vals.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)

	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("keycloak token request failed %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil

}

type KeycloakClient struct {
	ID                           string        `json:"id"`
	ClientID                     string        `json:"clientId"`
	Name                         string        `json:"name"`
	Description                  string        `json:"description"`
	SurrogateAuthRequired        bool          `json:"surrogateAuthRequired"`
	Enabled                      bool          `json:"enabled"`
	AlwaysDisplayInConsole       bool          `json:"alwaysDisplayInConsole"`
	ClientAuthenticatorType      string        `json:"clientAuthenticatorType"`
	Secret                       string        `json:"secret"`
	RedirectUris                 []interface{} `json:"redirectUris"`
	WebOrigins                   []interface{} `json:"webOrigins"`
	NotBefore                    int           `json:"notBefore"`
	BearerOnly                   bool          `json:"bearerOnly"`
	ConsentRequired              bool          `json:"consentRequired"`
	StandardFlowEnabled          bool          `json:"standardFlowEnabled"`
	ImplicitFlowEnabled          bool          `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled    bool          `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled       bool          `json:"serviceAccountsEnabled"`
	AuthorizationServicesEnabled bool          `json:"authorizationServicesEnabled"`
	PublicClient                 bool          `json:"publicClient"`
	FrontchannelLogout           bool          `json:"frontchannelLogout"`
	Protocol                     string        `json:"protocol"`
	Attributes                   struct {
		OidcCibaGrantEnabled                  string `json:"oidc.ciba.grant.enabled"`
		Oauth2DeviceAuthorizationGrantEnabled string `json:"oauth2.device.authorization.grant.enabled"`
		ClientSecretCreationTime              string `json:"client.secret.creation.time"`
		BackchannelLogoutSessionRequired      string `json:"backchannel.logout.session.required"`
		BackchannelLogoutRevokeOfflineTokens  string `json:"backchannel.logout.revoke.offline.tokens"`
	} `json:"attributes"`
	AuthenticationFlowBindingOverrides struct {
	} `json:"authenticationFlowBindingOverrides"`
	FullScopeAllowed          bool `json:"fullScopeAllowed"`
	NodeReRegistrationTimeout int  `json:"nodeReRegistrationTimeout"`
	ProtocolMappers           []struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Protocol        string `json:"protocol"`
		ProtocolMapper  string `json:"protocolMapper"`
		ConsentRequired bool   `json:"consentRequired"`
		Config          struct {
			UserSessionNote  string `json:"user.session.note"`
			IDTokenClaim     string `json:"id.token.claim"`
			AccessTokenClaim string `json:"access.token.claim"`
			ClaimName        string `json:"claim.name"`
			JSONTypeLabel    string `json:"jsonType.label"`
		} `json:"config"`
	} `json:"protocolMappers"`
	DefaultClientScopes  []string `json:"defaultClientScopes"`
	OptionalClientScopes []string `json:"optionalClientScopes"`
	Access               struct {
		View      bool `json:"view"`
		Configure bool `json:"configure"`
		Manage    bool `json:"manage"`
	} `json:"access"`
}

func (v *vaultAuth) getClientIDTokenFromKeycloak(ctx context.Context, httpClient *http.Client, adminToken string) (string, error) {
	keycloakURL := v.keycloakServer + "/admin/realms/master/clients?clientId=" + keycloakUserClientName

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		keycloakURL, nil,
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+adminToken)

	resp, err := httpClient.Do(req)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("keycloak client id request failed %d", resp.StatusCode)
	}

	if err != nil {
		return "", err
	}

	clientResp := []KeycloakClient{}

	err = json.NewDecoder(resp.Body).Decode(&clientResp)
	if err != nil {
		return "", err
	}

	clientID := clientResp[0].ID

	return clientID, nil
}

type KeycloakSecret struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (v *vaultAuth) getSecretFromKeycloak(ctx context.Context, httpClient *http.Client, clientID string, adminToken string) (string, error) {
	keycloakURL := v.keycloakServer + "/admin/realms/master/clients/" + clientID + "/client-secret"

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		keycloakURL, nil,
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+adminToken)

	resp, err := httpClient.Do(req)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("keycloak secret request failed %d", resp.StatusCode)
	}

	if err != nil {
		return "", err
	}

	secretResp := KeycloakSecret{}

	err = json.NewDecoder(resp.Body).Decode(&secretResp)
	if err != nil {
		return "", err
	}

	clientSecret := secretResp.Value

	return clientSecret, nil
}

// getClientSecretFromKeycloak reads and returns the client secret from the vault
func (v *vaultAuth) getClientSecretFromKeycloak(ctx context.Context, httpClient *http.Client, adminUsername string, adminPassword string) (string, string, error) {
	adminToken, err := v.getAdminTokenFromKeycloak(ctx, httpClient, adminUsername, adminPassword)
	if err != nil {
		return "", "", err
	}

	clientID, _ := v.getClientIDTokenFromKeycloak(ctx, httpClient, adminToken)

	secret, err := v.getSecretFromKeycloak(ctx, httpClient, clientID, adminToken)

	return clientID, secret, err
}

// setClientSecretToVault stores the client secret in the vault
func (v *vaultAuth) setClientSecretToVault(ctx context.Context, httpClient *http.Client, vaultClientToken string, adminUsername string, adminPassword string) (string, error) {
	clientID, clientSecret, err := v.getClientSecretFromKeycloak(ctx, httpClient, adminUsername, adminPassword)

	if err != nil {
		return "", err
	}
	vaultURL := v.httpsVaultURL(vaultSecretBaseURL + m2mVaultSecretPath)

	type VaultSecretDataData struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	type VaultSetSecret struct {
		Data VaultSecretDataData `json:"data"`
	}
	clientSecretData := VaultSetSecret{
		Data: VaultSecretDataData{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
	}

	configBody, err := json.Marshal(clientSecretData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		vaultURL,
		bytes.NewReader(configBody),
	)
	if err != nil {
		return "", err
	}

	req.Header.Add("X-Vault-Token", vaultClientToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	getRawData, err := readAllFactory(resp.Body)
	if err != nil {
		return "", err
	}

	var secret vault.Secret
	err = json.Unmarshal(getRawData, &secret)
	if err != nil {
		return "", err
	}

	type VaultPathGetResponse struct {
		Data struct {
			Data struct {
				ClientID     string `json:"client_id"`
				ClientSecret string `json:"client_secret"`
			} `json:"data"`
		} `json:"data"`
	}
	var getResult VaultPathGetResponse
	err = json.Unmarshal(getRawData, &getResult)
	if err != nil {
		return "", err
	}
	return getResult.Data.Data.ClientSecret, nil
}

// getM2MTokenFromKeycloak reads and returns the M2M token from keycloak
func (v *vaultAuth) getM2MTokenFromKeycloak(ctx context.Context, httpClient *http.Client, clientSecret string) (string, error) {
	keycloakURL := v.keycloakServer + keycloakTokenURL
	vals := url.Values{}
	vals.Add("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		keycloakURL, strings.NewReader(vals.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("edge-manager-m2m-client", clientSecret)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("keycloak m2m token request failed %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func (v *vaultAuth) GetM2MToken(ctx context.Context) (string, error) {
	useM2MTokenString := os.Getenv("USE_M2M_TOKEN")
	useM2MToken, err := strconv.ParseBool(useM2MTokenString)
	if err != nil || !useM2MToken {
		return "", nil
	}
	M2MToken := ""

	// get vault token
	vaultClient, err := getHTTPClient()
	if err != nil {
		return "", err
	}
	vaultClientToken, err := v.GetVaultToken(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = v.Logout(ctx) }()
	// get client name/client secret from vault
	clientSecret, err := v.getClientSecretFromVault(ctx, vaultClient, vaultClientToken)
	if err != nil {
		return "", err
	}

	// read M2M token
	M2MToken, err = v.getM2MTokenFromKeycloak(ctx, vaultClient, clientSecret)
	if err != nil {
		return "", err
	}

	return M2MToken, nil
}

func (v *vaultAuth) CreateClientSecret(ctx context.Context, username string, password string) (string, error) {
	useM2MTokenString := os.Getenv("USE_M2M_TOKEN")
	useM2MToken, err := strconv.ParseBool(useM2MTokenString)
	if err != nil || !useM2MToken {
		return "", nil
	}
	M2MToken := ""

	// get vault token
	vaultClient, err := getHTTPClient()
	if err != nil {
		return "", err
	}
	vaultClientToken, err := v.GetVaultToken(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = v.Logout(ctx) }()

	// set client name/client secret in vault
	M2MToken, err = v.setClientSecretToVault(ctx, vaultClient, vaultClientToken, username, password)
	if err != nil {
		return "", err
	}

	return M2MToken, nil
}

func (v *vaultAuth) Logout(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.vaultToken == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		v.httpsVaultURL(vaultRevokeSelfURL),
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", v.vaultToken)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		log.Infof("http error on revoke: %d", resp.StatusCode)
		return fmt.Errorf("http error on revoke: %d", resp.StatusCode)
	}
	v.vaultToken = ""
	return nil
}
