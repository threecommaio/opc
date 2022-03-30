// Package version handles embeding the version information in the binary
package version

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

// version vars
var (
	BuildRelease   = "project/v0.0.0"
	Project        = "project"
	Version        = "v0.0.0"
	CommitHash     = "UNKNOWN"
	BuildTimestamp = time.Now()
	DirtyBuild     = false
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			CommitHash = kv.Value
		case "vcs.time":
			BuildTimestamp, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			DirtyBuild = kv.Value == "true"
		}
	}
	r := strings.Split(BuildRelease, "/")
	if len(r) == 2 {
		Project = r[0]
		Version = r[1]
	}
}

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
	return BuildTimestamp.String()
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
