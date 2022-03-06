package web

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v43/github"
	"github.com/slack-go/slack"
)

const (
	SlackSignature  = "X-Slack-Signature"
	GithubSignature = "X-Hub-Signature-256"
	LinearSignature = "Linear-Delivery"
)

var (
	ErrVerifyingSlack  = errors.New("error verifying slack secret")
	ErrVerifyingGithub = errors.New("error verifying github secret")
)

type Signature struct {
	Header  string
	Env     string
	IsValid func(c *gin.Context, data []byte, secret string) error
}

// SlackValidator validates slack webhook signature
func SlackValidator(c *gin.Context, data []byte, secret string) error {
	sv, err := slack.NewSecretsVerifier(c.Request.Header, secret)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrVerifyingSlack, err)
	}
	_, err = sv.Write(data)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrVerifyingSlack, err)
	}
	err = sv.Ensure()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrVerifyingSlack, err)
	}
	return nil
}

// GithubValidator validates github webhook signature
func GithubValidator(c *gin.Context, data []byte, secret string) error {
	err := github.ValidateSignature(GithubSignature, data, []byte(secret))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrVerifyingGithub, err)
	}
	return nil
}

// LinearValidator validates linear webhook signature
func LinearValidator(c *gin.Context, data []byte, secret string) error {
	// TODO: add ip allowlist here
	// no verification required due to linear lacking signature verification
	return nil
}
