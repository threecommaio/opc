// Package core provides the core functionality for an application
package core

import (
	// force imports.
	_ "github.com/Masterminds/goutils"
	_ "github.com/cenkalti/backoff"
	_ "github.com/creasty/defaults"
	_ "github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	_ "github.com/google/uuid"
	_ "github.com/huandu/xstrings"
	_ "github.com/imdario/mergo"
	_ "github.com/jinzhu/now"
	_ "github.com/kr/pretty"
	"github.com/magefile/mage/sh"
	_ "github.com/mitchellh/copystructure"
	_ "github.com/muyo/sno"
	_ "github.com/olebedev/when"
	_ "github.com/olekukonko/tablewriter"
	_ "github.com/tidwall/jsonc"
	_ "go.uber.org/ratelimit"
	_ "gopkg.in/yaml.v3"
)

const (
	Production  = "production"
	Staging     = "staging"
	Development = "development"
)

// Environment is the environment for the application
func Environment() string {
	switch gin.Mode() {
	case gin.DebugMode:
		return Development
	case gin.TestMode:
		return Staging
	case gin.ReleaseMode:
		return Development
	}

	return "unknown"
}

// RunCmd runs a command
func RunCmd(cmd string, args ...string) func(args ...string) error {
	return func(args2 ...string) error {
		return sh.Run(cmd, append(args, args2...)...)
	}
}
