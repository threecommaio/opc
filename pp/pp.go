// Package pp provides helpers for pretty printing
package pp

import (
	"encoding/json"

	"github.com/tidwall/pretty"
)

// JSON converts a struct to a pretty-printed JSON string.
func JSON(message interface{}) string {
	b, _ := json.MarshalIndent(message, "", "  ")
	return string(b)
}

// PrettyJSON takes a JSON string and returns a pretty-printed version
func PrettyJSON(json []byte) string {
	return string(pretty.Pretty(json))
}
