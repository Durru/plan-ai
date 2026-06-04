// Package scanner provides deterministic, rule-based local project analysis.
//
// The scanner inspects a project root and produces a stable, testable snapshot
// of the project's git state, languages, frameworks, package managers,
// dependencies, and important files. It contains no AI, no network calls,
// and no external services. The output is suitable for downstream phases
// (planner, research, context engine) to consume as a stable substrate.
package scanner

import "time"

// FileKind is the high-level classification of a file used to build summary
// views of the project. The classification is intentionally coarse; finer
// details live in extensions and filenames.
type FileKind string

const (
	// FileKindSource marks typical source-code files.
	FileKindSource FileKind = "source"
	// FileKindConfig marks project configuration files (JSON, YAML, etc.).
	FileKindConfig FileKind = "config"
	// FileKindDoc marks documentation and prose files.
	FileKindDoc FileKind = "doc"
	// FileKindTest marks test files across supported languages.
	FileKindTest FileKind = "test"
	// FileKindLock marks lockfile artifacts produced by package managers.
	FileKindLock FileKind = "lock"
	// FileKindOther is the fallback for files that do not match a known
	// pattern but were still considered important.
	FileKindOther FileKind = "other"
)

// ScannedFile describes a single file the scanner considered.
type ScannedFile struct {
	// Path is the file path relative to the scanned project root, always
	// using forward slashes.
	Path string
	// Kind is the file's high-level classification.
	Kind FileKind
	// Size is the file's size in bytes at scan time. Files larger than the
	// scanner's size cap are never recorded.
	Size int64
}

// LanguageStats counts the number of indexed files per language.
type LanguageStats struct {
	Language string
	Files    int
}

// FrameworkHit records a framework the scanner detected and the evidence
// (file path or dependency entry) that triggered the detection.
type FrameworkHit struct {
	Name     string
	Evidence string
}

// PackageManagerHit records a package manager the scanner detected and the
// evidence file (lockfile, manifest) that triggered the detection.
type PackageManagerHit struct {
	Name     string
	Evidence string
}

// Dependency represents a single external library known to the project.
type Dependency struct {
	Name    string
	Version string
	Source  string
}

// Result is the full deterministic scan of a project root.
type Result struct {
	// ProjectRoot is the absolute path that was scanned.
	ProjectRoot string
	// GitDetected is true when a .git directory, worktree, or git binary
	// reported the project as inside a working tree.
	GitDetected bool
	// GitBranch is the current branch when one could be resolved; empty
	// otherwise.
	GitBranch string
	// Fingerprint is a stable hash of the scan inputs (files, sizes, mtimes,
	// project root). It is intended for change detection, not identity.
	Fingerprint string
	// Languages lists the languages detected, sorted by language name.
	Languages []LanguageStats
	// Frameworks lists the detected frameworks, deduplicated and sorted.
	Frameworks []FrameworkHit
	// PackageManagers lists the detected package managers, deduplicated and
	// sorted.
	PackageManagers []PackageManagerHit
	// Dependencies lists all parsed dependencies, sorted by source then name.
	Dependencies []Dependency
	// Files lists every file the scanner indexed, in a stable order.
	Files []ScannedFile
	// Summary is a one-line human-readable description of the scan.
	Summary string
	// CreatedAt is the moment the scan was produced.
	CreatedAt time.Time
}
