// Package version provides the Plan-AI build version, injected at build time
// via ldflags: -X github.com/Durru/plan-ai/internal/version.Version=1.0.0
package version

// Version is the current Plan-AI version. Set at build time via ldflags.
// Default "dev" means built from source without a release tag.
var Version = "dev"
