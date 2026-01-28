package version

import "fmt"

// These variables are intended to be set via -ldflags at build time.
//
// Example:
//   go build -ldflags "-X github.com/ikepcampbell/kubemedic/internal/version.Version=0.1.0 \
//     -X github.com/ikepcampbell/kubemedic/internal/version.GitCommit=$(git rev-parse --short HEAD) \
//     -X github.com/ikepcampbell/kubemedic/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	Version   = "dev"
	GitCommit = ""
	BuildDate = ""
)

func String() string {
	if GitCommit == "" && BuildDate == "" {
		return Version
	}
	return fmt.Sprintf("%s (commit=%s, date=%s)", Version, emptyAsUnknown(GitCommit), emptyAsUnknown(BuildDate))
}

func emptyAsUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}
