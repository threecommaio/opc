// Package oidc handles the logic for the OpenID Connect protocol across GCP
package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	htransport "google.golang.org/api/transport/http"
)

// Github Action OIDC environment variables
const (
	ActionsIDTokenRequestToken = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
	ActionsIDTokenRequestURL   = "ACTIONS_ID_TOKEN_REQUEST_URL"

	// Content Types
	ApplicationJSON = "application/json"

	// GOOGLE OIDC
	STSURL               = "https://sts.googleapis.com/v1/token"
	AuthGrantType        = "urn:ietf:params:oauth:grant-type:token-exchange"
	AuthAudience         = "//iam.googleapis.com/"
	AuthRequestTokenType = "urn:ietf:params:oauth:token-type:access_token"
	AuthTokenScope       = "https://www.googleapis.com/auth/cloud-platform"
	AuthSubjectTokenType = "urn:ietf:params:oauth:token-type:jwt"
)

// errors
var (
	ErrNoToken = errors.New("token did not contain an id_token")
)

// RequestAccessToken is the request payload for the OIDC token exchange
type RequestAccessToken struct {
	Audience           string `json:"audience"`
	GrantType          string `json:"grant_type"`
	RequestedTokenType string `json:"requested_token_type"`
	Scope              string `json:"scope"`
	SubjectToken       string `json:"subject_token"`
	SubjectTokenType   string `json:"subject_token_type"`
}

// TokenRequest is the token payload for the OIDC token exchange
type TokenRequest struct {
	Audience string `json:"audience,omitempty"`
	// IncludeEmail bool     `json:"includeEmail,omitempty"`
	Delegates []string `json:"delegates,omitempty"`
}

// WorkloadIdentityRequest is the workload identity payload for the OIDC token exchange
type WorkloadIdentityRequest struct {
	OIDCRequestToken string
	OIDCRequestURL   string
	// projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider
	WorkloadIdentityProvider string
	ServiceAccount           string // my-service-account@my-project.iam.gserviceaccount.com
	IDTokenAudience          string // https://demo-uc.a.run.app
}

// GetIDToken gets the ID token for the given service account
func GetIDToken(ctx context.Context, requestURL, requestTokn, providerID string) (string, error) {
	client := &http.Client{}
	audience := `https://iam.googleapis.com/` + providerID
	requestURL = requestURL + "&audience=" + audience

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+requestTokn)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get ID token: %w", err)
	}
	defer resp.Body.Close()
	var payload struct {
		Count int    `json:"count"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("failed to decode ID token: %w", err)
	}

	return payload.Value, nil
}

// GetAuthToken gets the auth token for the given service account
func GetAuthToken(ctx context.Context, providerID, token string) (string, error) {
	raToken := RequestAccessToken{
		Audience:           AuthAudience + providerID,
		GrantType:          AuthGrantType,
		RequestedTokenType: AuthRequestTokenType,
		Scope:              AuthTokenScope,
		SubjectTokenType:   AuthSubjectTokenType,
		SubjectToken:       token,
	}

	jsonBody, err := json.MarshalIndent(raToken, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal request access token: %w", err)
	}
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", STSURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Content-Type", ApplicationJSON)
	req.Header.Add("Accept", ApplicationJSON)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get auth token: %w", err)
	}

	var rat struct {
		AccessToken string `json:"access_token"`
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(bytes.NewBuffer(data)).Decode(&rat); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return rat.AccessToken, nil
}

// GoogleIDToken is the ID token for a Google service account
func GoogleIDToken(ctx context.Context, token, sa, audience string) (string, error) {
	serviceAccountID := `projects/-/serviceAccounts/` + sa
	tokenURL := `https://iamcredentials.googleapis.com/v1/` + serviceAccountID + `:generateIdToken`

	it := TokenRequest{
		Audience: audience,
	}

	jsonBody, err := json.MarshalIndent(it, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal id token: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", ApplicationJSON)
	req.Header.Add("Accept", ApplicationJSON)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get id token: %w", err)
	}

	var payload struct {
		Token string `json:"token,omitempty"`
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(bytes.NewBuffer(data)).Decode(&payload); err != nil {
		return "", fmt.Errorf("failed to decode response body: %w", err)
	}

	return payload.Token, nil
}

// WorkloadIdentityToken gets the ID token thats required to hit an authenticated endpoint
func WorkloadIdentityToken(ctx context.Context, req WorkloadIdentityRequest) (string, error) {
	oidcToken, err := GetIDToken(ctx, req.OIDCRequestURL, req.OIDCRequestToken,
		req.WorkloadIdentityProvider)
	if err != nil {
		return "", err
	}
	accessToken, err := GetAuthToken(ctx, req.WorkloadIdentityProvider, oidcToken)
	if err != nil {
		return "", err
	}
	idToken, err := GoogleIDToken(ctx, accessToken, req.ServiceAccount, req.IDTokenAudience)
	if err != nil {
		return "", err
	}

	return idToken, nil
}

// NewProxy takes target host and creates a reverse proxy
func NewProxy(ctx context.Context, targetHost string) (*httputil.ReverseProxy, error) {
	tr, err := NewTransport(ctx)
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target host: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = tr

	return proxy, nil
}

// idTokenSource is an oauth2.TokenSource that wraps another
// It takes the id_token from TokenSource and passes that on as a bearer token
type idTokenSource struct {
	TokenSource oauth2.TokenSource
}

// Token returns a token that is the same as the one returned by the wrapped TokenSource
func (s *idTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, ErrNoToken
	}

	return &oauth2.Token{
		AccessToken: idToken,
		TokenType:   "Bearer",
		Expiry:      token.Expiry,
	}, nil
}

// NewTransport returns a new transport that adds the id_token to the request
func NewTransport(ctx context.Context) (http.RoundTripper, error) {
	gts, err := google.DefaultTokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get default token source: %w", err)
	}
	ts := oauth2.ReuseTokenSource(nil, &idTokenSource{TokenSource: gts})

	opts := make([]option.ClientOption, 0, 2)
	opts = append(opts, option.WithTokenSource(ts), internaloption.SkipDialSettingsValidation())
	t, err := htransport.NewTransport(ctx, http.DefaultTransport, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	return t, nil
}
