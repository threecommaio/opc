package web

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v43/github"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
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
	// handle parsing slack events including challenge requests
	event, err := slackevents.ParseEvent(data, slackevents.OptionNoVerifyToken())
	if err != nil {
		return err
	}
	// challenge request for first time setting up the webhook in slack app
	if event.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal(data, &r)
		if err != nil {
			return fmt.Errorf("error parsing challenge response: %s", err)
		}
		c.Writer.Header().Set("Content-Type", "text")
		_, err = c.Writer.Write([]byte(r.Challenge))
		if err != nil {
			return fmt.Errorf("error writing challenge response: %s", err)
		}
		return nil
	}

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
	signature := c.Request.Header.Get(GithubSignature)
	err := github.ValidateSignature(signature, data, []byte(secret))
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
