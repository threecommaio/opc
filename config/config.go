// Package config handles parsing and validation of the configuration file
package config

import (
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v3"
)

const (
	defaultPath = "config.yaml"
)

// Config interface
type Config any

// Parse handles parsing the default location of the config file
func Parse(dest any, f fs.FS) (err error) {
	return ParseFile(dest, f, defaultPath)
}

// ParseFile handles parsing a config file and unmarshaling it into the dest
func ParseFile(dest any, f fs.FS, filename string) (err error) {
	// Open config file
	file, err := f.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&dest); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	return nil
}
