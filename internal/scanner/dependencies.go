package scanner

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Source identifiers used in Dependency.Source.
const (
	SourceGoMod        = "go.mod"
	SourcePackageJSON  = "package.json"
	SourceCargoToml    = "Cargo.toml"
	SourceRequirements = "requirements.txt"
	SourcePyproject    = "pyproject.toml"
	SourceComposer     = "composer.json"
	SourcePubspec      = "pubspec.yaml"
)

// ParseGoMod reads a go.mod file at the project root and returns the parsed
// require block as dependencies. Indirect dependencies are included with
// their versions when present. The caller is expected to have already
// verified the file is a regular file.
func ParseGoMod(root string) ([]Dependency, error) {
	path := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return parseGoModContent(string(data)), nil
}

func parseGoModContent(content string) []Dependency {
	seen := map[string]bool{}
	var out []Dependency
	inRequire := false
	scan := bufio.NewScanner(strings.NewReader(content))
	scan.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scan.Scan() {
		raw := scan.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "require ") {
			inRequire = true
			// Single-line require form: "require ( ... )" handled below.
			if strings.HasSuffix(strings.TrimSpace(line), ")") {
				inRequire = false
			}
			dep, ok := parseRequireLine(strings.TrimPrefix(line, "require "))
			if !ok {
				continue
			}
			if !seen[dep.Name] {
				seen[dep.Name] = true
				out = append(out, dep)
			}
			continue
		}
		if line == "require (" {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}
		if inRequire {
			dep, ok := parseRequireLine(line)
			if !ok {
				continue
			}
			if !seen[dep.Name] {
				seen[dep.Name] = true
				out = append(out, dep)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// requireLine matches the body of a go.mod require entry:
//
//	github.com/foo/bar v1.2.3
//	github.com/foo/bar v1.2.3 // indirect
var requireLine = regexp.MustCompile(`^([A-Za-z0-9._~+\-/]+)\s+([^\s]+)(\s+//\s*\w+.*)?$`)

func parseRequireLine(line string) (Dependency, bool) {
	// Strip the optional block prefix (e.g. "github.com/foo/bar v1.2.3 // indirect").
	stripped := strings.TrimSpace(line)
	matches := requireLine.FindStringSubmatch(stripped)
	if len(matches) < 3 {
		return Dependency{}, false
	}
	return Dependency{
		Name:    matches[1],
		Version: matches[2],
		Source:  SourceGoMod,
	}, true
}

// ParsePackageJSON reads package.json and returns the union of dependencies
// and devDependencies as Dependency records, sorted by name.
func ParsePackageJSON(root string) ([]Dependency, error) {
	path := filepath.Join(root, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var manifest struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}
	out := make([]Dependency, 0, len(manifest.Dependencies)+len(manifest.DevDependencies))
	for name, version := range manifest.Dependencies {
		out = append(out, Dependency{Name: name, Version: version, Source: SourcePackageJSON})
	}
	for name, version := range manifest.DevDependencies {
		out = append(out, Dependency{Name: name, Version: version, Source: SourcePackageJSON})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// ParseCargoToml reads Cargo.toml and returns a small best-effort set of
// dependencies. The parser is intentionally simple and returns an empty
// slice (with no error) when the format is not recognized.
func ParseCargoToml(root string) ([]Dependency, error) {
	path := filepath.Join(root, "Cargo.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return parseCargoTomlContent(string(data)), nil
}

func parseCargoTomlContent(content string) []Dependency {
	var out []Dependency
	seen := map[string]bool{}
	inTable := ""
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	cargoLine := regexp.MustCompile(`^([A-Za-z0-9_\-]+)\s*=\s*("?[^"]+?"?)`)
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
			inTable = strings.Trim(raw, "[]")
			continue
		}
		if inTable != "dependencies" && inTable != "dev-dependencies" {
			continue
		}
		matches := cargoLine.FindStringSubmatch(raw)
		if len(matches) < 2 {
			continue
		}
		name := matches[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, Dependency{Name: name, Version: strings.Trim(matches[2], "\""), Source: SourceCargoToml})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ParseRequirementsTxt reads a pip requirements file and returns one
// dependency per non-empty, non-comment line.
func ParseRequirementsTxt(root string) ([]Dependency, error) {
	path := filepath.Join(root, "requirements.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []Dependency
	seen := map[string]bool{}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip inline comments and trailing whitespace.
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}
		name, version := splitPipRequirement(line)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, Dependency{Name: name, Version: version, Source: SourceRequirements})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

var pipReq = regexp.MustCompile(`^([A-Za-z0-9._\-]+)\s*([=<>!~]=?\s*[^;\s]+)?`)

func splitPipRequirement(line string) (string, string) {
	matches := pipReq.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", ""
	}
	name := matches[1]
	version := ""
	if len(matches) >= 3 && matches[2] != "" {
		version = strings.TrimSpace(matches[2])
	}
	return name, version
}

// ParsePyprojectToml reads pyproject.toml and returns any PEP 621
// dependencies or [tool.poetry.dependencies] entries. The parser is
// intentionally lenient: unparsable files yield an empty slice with no
// error.
func ParsePyprojectToml(root string) ([]Dependency, error) {
	path := filepath.Join(root, "pyproject.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return parsePyprojectTomlContent(string(data)), nil
}

func parsePyprojectTomlContent(content string) []Dependency {
	var out []Dependency
	seen := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	inTable := false
	pyprojectLine := regexp.MustCompile(`^([A-Za-z0-9._\-]+)\s*=\s*("?[^"]+?"?)`)
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
			header := strings.Trim(raw, "[]")
			inTable = header == "project.dependencies" || header == "tool.poetry.dependencies" || header == "dependencies"
			continue
		}
		if !inTable {
			continue
		}
		matches := pyprojectLine.FindStringSubmatch(raw)
		if len(matches) < 2 {
			continue
		}
		name := matches[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, Dependency{Name: name, Version: strings.Trim(matches[2], "\""), Source: SourcePyproject})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ParseComposerJSON reads composer.json and returns require entries.
func ParseComposerJSON(root string) ([]Dependency, error) {
	path := filepath.Join(root, "composer.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var manifest struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse composer.json: %w", err)
	}
	out := make([]Dependency, 0, len(manifest.Require)+len(manifest.RequireDev))
	for name, version := range manifest.Require {
		// Skip the PHP platform requirement key, which is a "name" too.
		if name == "php" {
			continue
		}
		out = append(out, Dependency{Name: name, Version: version, Source: SourceComposer})
	}
	for name, version := range manifest.RequireDev {
		out = append(out, Dependency{Name: name, Version: version, Source: SourceComposer})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// ParsePubspecYaml reads pubspec.yaml and returns a single dependency for
// the project name when present, plus any flutter package entries. The
// parser is intentionally simple: it captures the top-level name/version
// only.
func ParsePubspecYaml(root string) ([]Dependency, error) {
	path := filepath.Join(root, "pubspec.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return parsePubspecContent(string(data)), nil
}

func parsePubspecContent(content string) []Dependency {
	var out []Dependency
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	pubspecLine := regexp.MustCompile(`^([A-Za-z0-9_]+):\s*(.*)$`)
	for scanner.Scan() {
		raw := scanner.Text()
		if strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t") {
			continue
		}
		matches := pubspecLine.FindStringSubmatch(strings.TrimSpace(raw))
		if len(matches) < 2 {
			continue
		}
		if matches[1] == "name" || matches[1] == "version" {
			if matches[1] == "name" {
				out = append(out, Dependency{Name: strings.Trim(matches[2], "\""), Source: SourcePubspec})
			}
		}
	}
	return out
}

// AllDependencies aggregates the per-ecosystem parsers. The returned slice
// is sorted by (source, name) and deduplicated by (name, source) pair.
func AllDependencies(root string) []Dependency {
	parsers := []func(string) ([]Dependency, error){
		ParseGoMod,
		ParsePackageJSON,
		ParseCargoToml,
		ParseRequirementsTxt,
		ParsePyprojectToml,
		ParseComposerJSON,
		ParsePubspecYaml,
	}
	var out []Dependency
	for _, parser := range parsers {
		deps, err := parser(root)
		if err != nil {
			continue
		}
		out = append(out, deps...)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].Name < out[j].Name
	})
	return dedupeDependencies(out)
}

func dedupeDependencies(in []Dependency) []Dependency {
	if len(in) == 0 {
		return in
	}
	seen := map[string]bool{}
	out := make([]Dependency, 0, len(in))
	for _, dep := range in {
		key := dep.Source + "::" + dep.Name
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, dep)
	}
	return out
}
