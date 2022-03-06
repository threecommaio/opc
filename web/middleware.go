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

// signatures we should verify if they exist
var validSignatures []Signature = []Signature{
	{
		Header:  SlackSignature,
		Env:     "SLACK_WEBHOOK_SECRET",
		IsValid: SlackValidator,
	},
	{
		Header:  GithubSignature,
		Env:     "GH_WEBHOOK_SECRET",
		IsValid: GithubValidator,
	},
	{
		Header:  LinearSignature,
		Env:     "LINEAR_WEBHOOK_SECRET", // not implemented yet
		IsValid: LinearValidator,
	},
}

func WebhookSecretValidation() gin.HandlerFunc {
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

		for _, signature := range validSignatures {
			_, ok := c.Request.Header[signature.Header]
			if ok {
				secret := os.Getenv(signature.Env)
				err := signature.IsValid(c, data, secret)
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

// IsSecretValidateEnabled checks if secret validation should be enabled
func IsSecretValidateEnabled() bool {
	return core.Environment() == core.Production
}
