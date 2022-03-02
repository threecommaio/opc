// Package core provides the core functionality for an application
package core

import (
	// force imports.
	_ "github.com/Masterminds/goutils"
	_ "github.com/cenkalti/backoff"
	_ "github.com/creasty/defaults"
	_ "github.com/dustin/go-humanize"
	_ "github.com/google/uuid"
	_ "github.com/huandu/xstrings"
	_ "github.com/imdario/mergo"
	_ "github.com/jinzhu/now"
	_ "github.com/kr/pretty"
	_ "github.com/mitchellh/copystructure"
	_ "github.com/muyo/sno"
	_ "github.com/olebedev/when"
	_ "github.com/olekukonko/tablewriter"
	_ "github.com/tidwall/jsonc"
	_ "go.uber.org/ratelimit"
	_ "gopkg.in/yaml.v3"
)
