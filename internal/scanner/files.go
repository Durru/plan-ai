package scanner

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultMaxFileSize is the size cap (1 MiB) above which the scanner refuses
// to record a file. Large binaries are unlikely to be source or config and
// would dominate the file index.
const DefaultMaxFileSize int64 = 1 << 20

// ignoredDirs is the set of directory base names the scanner never descends
// into. The match is case-sensitive on the final path segment.
var ignoredDirs = map[string]struct{}{
	".git":         {},
	".plan-ai":     {},
	"node_modules": {},
	"vendor":       {},
	"dist":         {},
	"build":        {},
	"coverage":     {},
	".tmp":         {},
	"tmp":          {},
	".cache":       {},
}

// ImportantFileKinds maps specific filenames to the high-level FileKind
// the scanner should use when other rules (extension, lockfile) do not apply.
var ImportantFileKinds = map[string]FileKind{
	"README.md":          FileKindDoc,
	"LICENSE":            FileKindDoc,
	"go.mod":             FileKindConfig,
	"go.sum":             FileKindLock,
	"package.json":       FileKindConfig,
	"pnpm-lock.yaml":     FileKindLock,
	"yarn.lock":          FileKindLock,
	"package-lock.json":  FileKindLock,
	"Cargo.toml":         FileKindConfig,
	"Cargo.lock":         FileKindLock,
	"requirements.txt":   FileKindConfig,
	"pyproject.toml":     FileKindConfig,
	"poetry.lock":        FileKindLock,
	"composer.json":      FileKindConfig,
	"pubspec.yaml":       FileKindConfig,
	"Dockerfile":         FileKindConfig,
	"docker-compose.yml": FileKindConfig,
	".env.example":       FileKindConfig,
	"Makefile":           FileKindConfig,
	"Taskfile.yml":       FileKindConfig,
	"justfile":           FileKindConfig,
}

// KindForPath returns the FileKind classification for the given relative
// path. The path must use forward slashes (e.g. "src/main.go").
func KindForPath(path string) FileKind {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	// 1. Lockfile match by exact filename (covers binaries like bun.lockb too).
	if isLockfileName(base) {
		return FileKindLock
	}

	// 2. Test-file match by filename pattern before extension checks.
	if isTestFileName(base) {
		return FileKindTest
	}

	// 3. Exact filename rules for important project files.
	if kind, ok := ImportantFileKinds[base]; ok {
		return kind
	}

	// 4. Source-code extensions.
	if isSourceExt(ext) {
		return FileKindSource
	}

	// 5. Common config / data formats.
	switch ext {
	case ".json", ".yaml", ".yml", ".toml", ".env":
		return FileKindConfig
	}

	// 6. Documentation.
	if ext == ".md" {
		return FileKindDoc
	}

	return FileKindOther
}

func isLockfileName(base string) bool {
	switch base {
	case "package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		"go.sum",
		"Cargo.lock",
		"poetry.lock",
		"bun.lockb",
		"composer.lock":
		return true
	}
	return false
}

func isTestFileName(base string) bool {
	lower := strings.ToLower(base)
	if strings.HasSuffix(lower, "_test.go") {
		return true
	}
	if strings.HasSuffix(lower, "_test.py") {
		return true
	}
	if strings.HasSuffix(lower, ".test.ts") || strings.HasSuffix(lower, ".test.tsx") {
		return true
	}
	if strings.HasSuffix(lower, ".spec.ts") || strings.HasSuffix(lower, ".spec.tsx") {
		return true
	}
	if strings.HasSuffix(lower, ".test.js") || strings.HasSuffix(lower, ".test.jsx") {
		return true
	}
	return false
}

func isSourceExt(ext string) bool {
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx",
		".py", ".rs", ".php", ".java", ".cs",
		".dart", ".sh", ".sql":
		return true
	}
	return false
}

// WalkProject returns the deterministic list of ScannedFile entries for the
// project rooted at root. The returned paths are relative to root and use
// forward slashes. Directories in ignoredDirs are skipped; files larger
// than maxSize are excluded.
func WalkProject(root string) ([]ScannedFile, error) {
	return walkProjectWithLimit(root, DefaultMaxFileSize)
}

func walkProjectWithLimit(root string, maxSize int64) ([]ScannedFile, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var out []ScannedFile
	walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(absRoot, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}

		if d.IsDir() {
			if _, skip := ignoredDirs[d.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if info.Size() > maxSize {
			return nil
		}

		relPath := filepath.ToSlash(rel)
		out = append(out, ScannedFile{
			Path: relPath,
			Kind: KindForPath(relPath),
			Size: info.Size(),
		})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out, nil
}
