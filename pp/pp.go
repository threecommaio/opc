// Package pp provides helpers for pretty printing
package pp

import "encoding/json"

// JSON converts a struct to a pretty-printed JSON string.
func JSON(message interface{}) string {
	b, _ := json.MarshalIndent(message, "", "  ")
	return string(b)
}
