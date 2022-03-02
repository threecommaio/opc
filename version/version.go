// Package version handles embeding the version information in the binary
package version

import (
	"fmt"
	"os"
	"path/filepath"
)

// version vars
var (
	Project        = "UNKNOWN"
	Version        = "v0.0.0"
	CommitHash     = "UNKNOWN"
	BuildTimestamp = "UNKNOWN"
)

// BuildVersion returns the full version of the build
func BuildVersion() string {
	return fmt.Sprintf("%s-%s (%s)", Version, CommitHash, BuildTimestamp)
}

// BuildVersionShort returns the short version of the build version
func BuildVersionShort() string {
	// major.minor.patch, includes v prefix
	return Version
}

// BuildFromCI returns true or false if built from CI
func BuildFromCI() bool {
	return CommitHash != "UNKNOWN"
}

// BuildTimestampFromCI returns the build timestamp from the CI environment
func BuildTimestampFromCI() string {
	return BuildTimestamp
}

// BinaryPath returns the path to the binary depending on environment
func BinaryPath() (string, error) {
	if BuildFromCI() {
		p, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("failed to get executable path: %w", err)
		}

		return filepath.Dir(p), nil
	}
	p, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return p, nil
}

// Release returns the release in the form of project@v1.0.0
func Release() string {
	// project@v1.0.0
	return fmt.Sprintf("%s@%s", Project, BuildVersionShort())
}
