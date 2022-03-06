package web

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v43/github"
	"github.com/slack-go/slack"
)

var (
	ErrVerifyingSlack  = errors.New("error verifying slack secret")
	ErrVerifyingGithub = errors.New("error verifying github secret")
)

type Signature struct {
	Header string
	Env    string
}

// SignatureValidator is a wrapper for signature validation
type SignatureValidator struct {
}

// IsSlackValid validates slack webhook signature
func (s *SignatureValidator) IsSlackValid(c *gin.Context, data []byte, secret string) error {
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

// IsGithubValid validates github webhook signature
func (s *SignatureValidator) IsGithubValid(c *gin.Context, data []byte, secret string) error {
	err := github.ValidateSignature(GithubSignature, data, []byte(secret))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrVerifyingGithub, err)
	}
	return nil
}

func NewSignatureValidator() *SignatureValidator {
	return &SignatureValidator{}
}
