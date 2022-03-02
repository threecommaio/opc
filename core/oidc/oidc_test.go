package oidc_test

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/threecommaio/opc/core/oidc"
)

const (
	WorkloadIdentityProvider = "projects/123456789/locations/global/" +
		"workloadIdentityPools/my-pool/providers/my-provider"
	ServiceAccount = "my-service-account@my-project.iam.gserviceaccount.com"
)

// Obtain the OIDC Token from Github Actions
func Example_oIDCToken() {
	oidcRequestURL := os.Getenv(oidc.ActionsIDTokenRequestURL)
	oidcRequestToken := os.Getenv(oidc.ActionsIDTokenRequestToken)
	oidcToken, err := oidc.GetIDToken(context.Background(), oidcRequestURL, oidcRequestToken,
		WorkloadIdentityProvider)
	if err != nil {
		panic(err)
	}
	fmt.Printf("OIDC Token: %s\n", oidcToken)
}

// Generate ID Token that can be used to authenticate against a CloudRun service
func Example_cloudRun() {
	oidcRequestURL := os.Getenv(oidc.ActionsIDTokenRequestURL)
	oidcRequestToken := os.Getenv(oidc.ActionsIDTokenRequestToken)

	idTokenAudience := "https://helloworld-snjhz2q4pa-uc.a.run.app"
	wreq := oidc.WorkloadIdentityRequest{
		OIDCRequestURL:           oidcRequestURL,
		OIDCRequestToken:         oidcRequestToken,
		WorkloadIdentityProvider: WorkloadIdentityProvider,
		ServiceAccount:           ServiceAccount,
		IDTokenAudience:          idTokenAudience,
	}
	token, err := oidc.WorkloadIdentityToken(context.Background(), wreq)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET",
		idTokenAudience+"/api/v1/hello", nil)
	if err != nil {
		panic(err)
	}

	// add authorization header to the req
	req.Header.Add("Authorization", "Bearer "+token)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
