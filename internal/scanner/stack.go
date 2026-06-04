package scanner

import (
	"path/filepath"
	"sort"
	"strings"
)

// LanguageExtensions maps a lower-case file extension (including the leading
// dot) to the human-friendly language name used in scan output.
var LanguageExtensions = map[string]string{
	".go":   "Go",
	".ts":   "TypeScript",
	".tsx":  "TypeScript",
	".js":   "JavaScript",
	".jsx":  "JavaScript",
	".py":   "Python",
	".rs":   "Rust",
	".php":  "PHP",
	".java": "Java",
	".cs":   "C#",
	".dart": "Dart",
	".md":   "Markdown",
	".sql":  "SQL",
	".sh":   "Shell",
	".yml":  "YAML",
	".yaml": "YAML",
	".json": "JSON",
	".toml": "TOML",
}

// PackageManagerFiles maps a project filename (no directory) to the
// (manager, evidence) tuple the scanner should record when that file is
// present.
var PackageManagerFiles = map[string]struct {
	Manager  string
	Evidence string
}{
	"go.mod":            {Manager: "go", Evidence: "go.mod"},
	"package-lock.json": {Manager: "npm", Evidence: "package-lock.json"},
	"pnpm-lock.yaml":    {Manager: "pnpm", Evidence: "pnpm-lock.yaml"},
	"yarn.lock":         {Manager: "yarn", Evidence: "yarn.lock"},
	"bun.lockb":         {Manager: "bun", Evidence: "bun.lockb"},
	"Cargo.toml":        {Manager: "cargo", Evidence: "Cargo.toml"},
	"requirements.txt":  {Manager: "pip", Evidence: "requirements.txt"},
	"pyproject.toml":    {Manager: "pip", Evidence: "pyproject.toml"},
	"poetry.lock":       {Manager: "poetry", Evidence: "poetry.lock"},
	"composer.json":     {Manager: "composer", Evidence: "composer.json"},
	"pubspec.yaml":      {Manager: "flutter", Evidence: "pubspec.yaml"},
}

// DetectLanguages tallies the file extensions in files and returns one
// LanguageStats entry per language, sorted by language name. Files with
// unknown extensions are skipped.
func DetectLanguages(files []ScannedFile) []LanguageStats {
	counts := map[string]int{}
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.Path))
		lang, ok := LanguageExtensions[ext]
		if !ok {
			continue
		}
		counts[lang]++
	}
	if len(counts) == 0 {
		return []LanguageStats{}
	}
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]LanguageStats, 0, len(names))
	for _, name := range names {
		out = append(out, LanguageStats{Language: name, Files: counts[name]})
	}
	return out
}

// DetectPackageManagers inspects the file map and returns the sorted list
// of detected package managers.
func DetectPackageManagers(fileMap map[string]bool) []PackageManagerHit {
	seen := map[string]bool{}
	var hits []PackageManagerHit
	for name, info := range PackageManagerFiles {
		if !fileMap[name] {
			continue
		}
		if seen[info.Manager] {
			continue
		}
		seen[info.Manager] = true
		hits = append(hits, PackageManagerHit{Name: info.Manager, Evidence: info.Evidence})
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Name < hits[j].Name })
	return hits
}

// FrameworkDetector inspects the collected project evidence and returns the
// frameworks it can recognize from that evidence. Implementations must be
// deterministic and side-effect free.
type FrameworkDetector struct {
	Name          string
	Detect        func(fileMap map[string]bool, packageJSONDeps, goModDeps map[string]string, otherFiles map[string]string) bool
	BuildEvidence func(fileMap map[string]bool, packageJSONDeps, goModDeps map[string]string, otherFiles map[string]string) string
}

// FrameworkDetectors is the ordered list of detectors the scanner applies.
// Order is informational only; outputs are deduplicated by name.
var FrameworkDetectors = []FrameworkDetector{
	{
		Name: "Next.js",
		Detect: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			return fileMap["next.config.js"] || fileMap["next.config.mjs"] || fileMap["next.config.ts"] || pkg["next"] != ""
		},
		BuildEvidence: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			for _, candidate := range []string{"next.config.js", "next.config.mjs", "next.config.ts"} {
				if fileMap[candidate] {
					return candidate
				}
			}
			if pkg["next"] != "" {
				return "package.json:next"
			}
			return "next.config"
		},
	},
	{
		Name: "React",
		Detect: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			_, ok := pkg["react"]
			return ok
		},
		BuildEvidence: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			return "package.json:react"
		},
	},
	{
		Name: "Vue",
		Detect: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			_, ok := pkg["vue"]
			return ok
		},
		BuildEvidence: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			return "package.json:vue"
		},
	},
	{
		Name: "Svelte",
		Detect: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			_, ok := pkg["svelte"]
			return ok
		},
		BuildEvidence: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			return "package.json:svelte"
		},
	},
	{
		Name: "Nuxt",
		Detect: func(fileMap map[string]bool, _, _ map[string]string, _ map[string]string) bool {
			return fileMap["nuxt.config.ts"] || fileMap["nuxt.config.js"]
		},
		BuildEvidence: func(fileMap map[string]bool, _, _ map[string]string, _ map[string]string) string {
			for _, candidate := range []string{"nuxt.config.ts", "nuxt.config.js"} {
				if fileMap[candidate] {
					return candidate
				}
			}
			return "nuxt.config"
		},
	},
	{
		Name: "Vite",
		Detect: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			return fileMap["vite.config.ts"] || fileMap["vite.config.js"] || pkg["vite"] != ""
		},
		BuildEvidence: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			for _, candidate := range []string{"vite.config.ts", "vite.config.js"} {
				if fileMap[candidate] {
					return candidate
				}
			}
			return "package.json:vite"
		},
	},
	{
		Name: "Express",
		Detect: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			return pkg["express"] != ""
		},
		BuildEvidence: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			return "package.json:express"
		},
	},
	{
		Name: "Fastify",
		Detect: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			return pkg["fastify"] != ""
		},
		BuildEvidence: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			return "package.json:fastify"
		},
	},
	{
		Name: "NestJS",
		Detect: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			return pkg["@nestjs/core"] != ""
		},
		BuildEvidence: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			return "package.json:@nestjs/core"
		},
	},
	{
		Name: "Go",
		Detect: func(fileMap map[string]bool, _, _ map[string]string, _ map[string]string) bool {
			return fileMap["go.mod"]
		},
		BuildEvidence: func(_ map[string]bool, _, _ map[string]string, _ map[string]string) string {
			return "go.mod"
		},
	},
	{
		Name: "Cobra",
		Detect: func(_ map[string]bool, _, gomod map[string]string, _ map[string]string) bool {
			return gomod["github.com/spf13/cobra"] != ""
		},
		BuildEvidence: func(_ map[string]bool, _, gomod map[string]string, _ map[string]string) string {
			return "go.mod:github.com/spf13/cobra"
		},
	},
	{
		Name: "SQLite",
		Detect: func(_ map[string]bool, _, gomod map[string]string, _ map[string]string) bool {
			return gomod["modernc.org/sqlite"] != "" || gomod["github.com/mattn/go-sqlite3"] != ""
		},
		BuildEvidence: func(_ map[string]bool, _, gomod map[string]string, _ map[string]string) string {
			if gomod["modernc.org/sqlite"] != "" {
				return "go.mod:modernc.org/sqlite"
			}
			return "go.mod:github.com/mattn/go-sqlite3"
		},
	},
	{
		Name: "MCP",
		Detect: func(_ map[string]bool, pkg, gomod map[string]string, _ map[string]string) bool {
			for dep := range gomod {
				if strings.Contains(dep, "mcp") {
					return true
				}
			}
			for dep := range pkg {
				if strings.Contains(dep, "mcp") {
					return true
				}
			}
			return false
		},
		BuildEvidence: func(_ map[string]bool, pkg, gomod map[string]string, _ map[string]string) string {
			for dep := range gomod {
				if strings.Contains(dep, "mcp") {
					return "go.mod:" + dep
				}
			}
			for dep := range pkg {
				if strings.Contains(dep, "mcp") {
					return "package.json:" + dep
				}
			}
			return "mcp"
		},
	},
	{
		Name: "Flutter",
		Detect: func(_ map[string]bool, _, _ map[string]string, other map[string]string) bool {
			content, ok := other["pubspec.yaml"]
			if !ok {
				return false
			}
			return strings.Contains(content, "flutter:")
		},
		BuildEvidence: func(_ map[string]bool, _, _ map[string]string, _ map[string]string) string {
			return "pubspec.yaml:flutter"
		},
	},
	{
		Name: "Supabase",
		Detect: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			return pkg["@supabase/supabase-js"] != "" || fileMap["supabase/config.toml"] || pathExistsUnderRoot(fileMap, "supabase")
		},
		BuildEvidence: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			if pkg["@supabase/supabase-js"] != "" {
				return "package.json:@supabase/supabase-js"
			}
			if fileMap["supabase/config.toml"] {
				return "supabase/config.toml"
			}
			return "supabase/"
		},
	},
	{
		Name: "Prisma",
		Detect: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			return fileMap["prisma/schema.prisma"] || pkg["prisma"] != ""
		},
		BuildEvidence: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			if fileMap["prisma/schema.prisma"] {
				return "prisma/schema.prisma"
			}
			return "package.json:prisma"
		},
	},
	{
		Name: "Drizzle",
		Detect: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			return pkg["drizzle-orm"] != ""
		},
		BuildEvidence: func(_ map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			return "package.json:drizzle-orm"
		},
	},
	{
		Name: "Tailwind",
		Detect: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) bool {
			return fileMap["tailwind.config.js"] || fileMap["tailwind.config.ts"] || pkg["tailwindcss"] != ""
		},
		BuildEvidence: func(fileMap map[string]bool, pkg, _ map[string]string, _ map[string]string) string {
			if fileMap["tailwind.config.js"] {
				return "tailwind.config.js"
			}
			if fileMap["tailwind.config.ts"] {
				return "tailwind.config.ts"
			}
			return "package.json:tailwindcss"
		},
	},
}

// pathExistsUnderRoot returns true if any file in fileMap begins with the
// given directory prefix.
func pathExistsUnderRoot(fileMap map[string]bool, dir string) bool {
	prefix := dir + "/"
	for path := range fileMap {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// DetectFrameworks applies the registered FrameworkDetectors to the
// collected project evidence and returns the deduplicated, sorted list of
// detected frameworks.
func DetectFrameworks(fileMap map[string]bool, packageJSONDeps, goModDeps map[string]string, otherFiles map[string]string) []FrameworkHit {
	seen := map[string]bool{}
	var hits []FrameworkHit
	for _, detector := range FrameworkDetectors {
		if seen[detector.Name] {
			continue
		}
		if !detector.Detect(fileMap, packageJSONDeps, goModDeps, otherFiles) {
			continue
		}
		seen[detector.Name] = true
		var evidence string
		if detector.BuildEvidence != nil {
			evidence = detector.BuildEvidence(fileMap, packageJSONDeps, goModDeps, otherFiles)
		}
		hits = append(hits, FrameworkHit{Name: detector.Name, Evidence: evidence})
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Name < hits[j].Name })
	return hits
}
