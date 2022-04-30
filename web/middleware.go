// middleware for handling webhook signature verification
package web

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/threecommaio/opc/core"
)

var (
	SlackSignature  = Signature{Header: HeaderSlack, Env: "SLACK_WEBHOOK_SECRET", IsValid: SlackValidator}
	SlackChallenge  = Challenge{Header: HeaderSlack, Env: "SLACK_WEBHOOK_SECRET", IsValid: SlackChallengeValidator}
	GithubSignature = Signature{Header: HeaderSlack, Env: "GH_WEBHOOK_SECRET", IsValid: GithubValidator}
	LinearSignature = Signature{Header: HeaderSlack, Env: "LINEAR_WEBHOOK_SECRET", IsValid: LinearValidator}
	// allSignatures we should verify if they exist
	allSignatures []Signature = []Signature{SlackSignature, GithubSignature, LinearSignature}
	allChallenges []Challenge = []Challenge{SlackChallenge}
)

// errors
var (
	ErrNoSignature = errors.New("no signature found")
)

// WebhookChallenge is a middleware for validating webhook challenge
func WebhookChallenge(opts ...Challenge) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := c.GetRawData()
		if IsError(c, err) {
			return
		}
		if opts == nil {
			opts = append(opts, allChallenges...)
		}

		for _, challenge := range opts {
			_, ok := c.Request.Header[challenge.Header]
			if ok {
				secret := os.Getenv(challenge.Env)
				err := challenge.IsValid(c, data, secret)
				if IsError401(c, err) {
					return
				}
				break
			}
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		c.Next()
	}
}

// WebhookSecretValidation is a middleware for validating webhook signature
func WebhookSecretValidation(opts ...Signature) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := c.GetRawData()
		if IsError(c, err) {
			return
		}

		// skip validation if not in production
		if !IsSecretValidateEnabled() {
			log.Warning("skipping secret validation not in production")
			c.Next()
			return
		}

		if opts == nil {
			opts = append(opts, allSignatures...)
		}
		valid := false
		for _, signature := range opts {
			_, ok := c.Request.Header[signature.Header]
			if ok {
				secret := os.Getenv(signature.Env)
				err := signature.IsValid(c, data, secret)
				if IsError401(c, err) {
					return
				}
				valid = true
				break
			}
		}
		// if no signature found, return error
		if !valid && IsError401(c, ErrNoSignature) {
			return
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		c.Next()
	}
}

// IsSecretValidateEnabled checks if secret validation should be enabled
func IsSecretValidateEnabled() bool {
	return core.Environment() == core.Production
}
