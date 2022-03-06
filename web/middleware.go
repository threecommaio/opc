// middleware for handling webhook signature verification
package web

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/threecommaio/opc/core"
)

const (
	SlackSignature  = "X-Slack-Signature"
	GithubSignature = "X-Hub-Signature-256"
	LinearSignature = "Linear-Delivery"
)

// signatures we should verify if they exist
var validSignatures []Signature = []Signature{
	{
		Header: SlackSignature,
		Env:    "SLACK_WEBHOOK_SECRET",
	},
	{
		Header: GithubSignature,
		Env:    "GH_WEBHOOK_SECRET",
	},
	{
		Header: LinearSignature,
		Env:    "LINEAR_WEBHOOK_SECRET", // not implemented yet
	},
}

func WebhookSecretValidation() gin.HandlerFunc {
	sv := NewSignatureValidator()
	return func(c *gin.Context) {
		// skip validation if not in production
		if !IsSecretValidateEnabled() {
			log.Warning("skipping secret validation not in production")
			c.Next()
			return
		}
		data, err := c.GetRawData()
		if IsError(c, err) {
			return
		}

	signature:
		for _, signature := range validSignatures {
			_, ok := c.Request.Header[signature.Header]
			if ok {
				secret := os.Getenv(signature.Env)
				switch signature.Header {
				case SlackSignature:
					err := sv.IsSlackValid(c, data, secret)
					if IsError401(c, err) {
						return
					}
					break signature
				case GithubSignature:
					err := sv.IsGithubValid(c, data, secret)
					if IsError401(c, err) {
						return
					}
					break signature
				case LinearSignature:
					// TODO: add ip allowlist here
					// no verification required due to linear lacking signature verification
					break signature
				}
			}
		}
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		c.Next()
	}
}

// IsSecretValidateEnabled checks if secret validation should be enabled
func IsSecretValidateEnabled() bool {
	return core.Environment() == core.Production
}
