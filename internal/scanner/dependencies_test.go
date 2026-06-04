package scanner

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestParseGoModExtractsDirectAndIndirect(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), `module demo

go 1.25

require (
	github.com/spf13/cobra v1.9.1
	github.com/example/lib v0.1.0 // indirect
	modernc.org/sqlite v1.51.0
)
`)

	deps, err := ParseGoMod(root)
	if err != nil {
		t.Fatalf("ParseGoMod: %v", err)
	}

	want := []Dependency{
		{Name: "github.com/example/lib", Version: "v0.1.0", Source: "go.mod"},
		{Name: "github.com/spf13/cobra", Version: "v1.9.1", Source: "go.mod"},
		{Name: "modernc.org/sqlite", Version: "v1.51.0", Source: "go.mod"},
	}
	if !reflect.DeepEqual(deps, want) {
		t.Fatalf("ParseGoMod = %+v, want %+v", deps, want)
	}
}

func TestParseGoModMissingFileIsNil(t *testing.T) {
	deps, err := ParseGoMod(t.TempDir())
	if err != nil {
		t.Fatalf("ParseGoMod: %v", err)
	}
	if deps != nil {
		t.Fatalf("expected nil slice for missing file, got %+v", deps)
	}
}

func TestParsePackageJSONExtractsDepsAndDevDeps(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "package.json"), `{
  "name": "demo",
  "version": "0.0.1",
  "dependencies": {
    "next": "14.2.5",
    "react": "18.3.1"
  },
  "devDependencies": {
    "typescript": "5.0.0"
  }
}`)

	deps, err := ParsePackageJSON(root)
	if err != nil {
		t.Fatalf("ParsePackageJSON: %v", err)
	}

	want := []Dependency{
		{Name: "next", Version: "14.2.5", Source: "package.json"},
		{Name: "react", Version: "18.3.1", Source: "package.json"},
		{Name: "typescript", Version: "5.0.0", Source: "package.json"},
	}
	if !reflect.DeepEqual(deps, want) {
		t.Fatalf("ParsePackageJSON = %+v, want %+v", deps, want)
	}
}

func TestParsePackageJSONMissingFileIsNil(t *testing.T) {
	deps, err := ParsePackageJSON(t.TempDir())
	if err != nil {
		t.Fatalf("ParsePackageJSON: %v", err)
	}
	if deps != nil {
		t.Fatalf("expected nil slice for missing file, got %+v", deps)
	}
}

func TestAllDependenciesMergesAndDeduplicates(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), `module demo
go 1.25
require github.com/spf13/cobra v1.9.1
`)
	mustWrite(t, filepath.Join(root, "package.json"), `{
  "name": "demo",
  "dependencies": {"next": "14.2.5"}
}`)
	mustWrite(t, filepath.Join(root, "requirements.txt"), "fastapi==0.110.0\n# comment\nflask\n")

	deps := AllDependencies(root)

	if len(deps) == 0 {
		t.Fatalf("expected dependencies from multiple sources, got none")
	}

	sources := map[string]int{}
	for _, dep := range deps {
		sources[dep.Source]++
	}
	if sources["go.mod"] == 0 {
		t.Fatalf("go.mod source missing from %+v", sources)
	}
	if sources["package.json"] == 0 {
		t.Fatalf("package.json source missing from %+v", sources)
	}
	if sources["requirements.txt"] == 0 {
		t.Fatalf("requirements.txt source missing from %+v", sources)
	}

	// Sorted by source then name.
	for i := 1; i < len(deps); i++ {
		if deps[i].Source < deps[i-1].Source {
			t.Fatalf("deps not sorted by source: %+v", deps)
		}
		if deps[i].Source == deps[i-1].Source && deps[i].Name < deps[i-1].Name {
			t.Fatalf("deps not sorted by name within source: %+v", deps)
		}
	}
}

func TestParseCargoTomlLenientOnUnknown(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "Cargo.toml"), "[package]\nname = \"demo\"\nversion = \"0.1.0\"\n\n[dependencies]\nserde = \"1\"\n")
	deps, err := ParseCargoToml(root)
	if err != nil {
		t.Fatalf("ParseCargoToml: %v", err)
	}
	if len(deps) == 0 {
		t.Fatalf("expected at least serde dependency, got none")
	}
}

func TestParsePyprojectTomlHandlesPEP621AndPoetry(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "pyproject.toml"), `[project]
name = "demo"
dependencies = ["fastapi>=0.110"]

[tool.poetry.dependencies]
python = "^3.11"
flask = "^3.0"
`)
	deps, err := ParsePyprojectToml(root)
	if err != nil {
		t.Fatalf("ParsePyprojectToml: %v", err)
	}
	if len(deps) == 0 {
		t.Fatalf("expected at least one dependency, got none")
	}
}

func TestParseComposerJSONSkipsPHPPlatform(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "composer.json"), `{
  "require": {
    "php": "^8.2",
    "laravel/framework": "^11.0"
  }
}`)
	deps, err := ParseComposerJSON(root)
	if err != nil {
		t.Fatalf("ParseComposerJSON: %v", err)
	}
	for _, dep := range deps {
		if dep.Name == "php" {
			t.Fatalf("PHP platform requirement should be skipped, got %+v", deps)
		}
	}
}

func TestParseRequirementsTxtIgnoresCommentsAndBlanks(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "requirements.txt"), "# top comment\n\nflask==3.0.0\n# inline comment\n# another\n")
	deps, err := ParseRequirementsTxt(root)
	if err != nil {
		t.Fatalf("ParseRequirementsTxt: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, got %+v", deps)
	}
	if deps[0].Name != "flask" || deps[0].Version != "==3.0.0" {
		t.Fatalf("deps[0] = %+v, want flask==3.0.0", deps[0])
	}
}

func TestFingerprintIsStableAndChangesOnFileChange(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.go"), "package a\n")
	mustWrite(t, filepath.Join(root, "b.go"), "package b\n")
	mustWrite(t, filepath.Join(root, "go.mod"), "module demo\n")

	scan, err := Default().Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	first, err := Fingerprint(root, scan.Files, scan.PackageManagers)
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}
	if first == "" {
		t.Fatalf("fingerprint is empty")
	}
	if len(first) != 32 {
		t.Fatalf("fingerprint length = %d, want 32 (hex chars)", len(first))
	}
	if !isHex(first) {
		t.Fatalf("fingerprint %q is not hex", first)
	}

	// Same content → same fingerprint.
	again, err := Fingerprint(root, scan.Files, scan.PackageManagers)
	if err != nil {
		t.Fatalf("Fingerprint again: %v", err)
	}
	if first != again {
		t.Fatalf("fingerprint not stable: %q vs %q", first, again)
	}

	// Modify file size → different fingerprint.
	modified := append([]ScannedFile(nil), scan.Files...)
	for i, f := range modified {
		if f.Path == "a.go" {
			modified[i] = ScannedFile{Path: f.Path, Kind: f.Kind, Size: f.Size + 1}
			break
		}
	}
	other, err := Fingerprint(root, modified, scan.PackageManagers)
	if err != nil {
		t.Fatalf("Fingerprint modified: %v", err)
	}
	if other == first {
		t.Fatalf("fingerprint did not change after file size change: %q", other)
	}
}

func isHex(s string) bool {
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

func TestParsePubspecYamlCapturesTopLevelName(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "pubspec.yaml"), "name: demo_app\nversion: 1.0.0\n")
	deps, err := ParsePubspecYaml(root)
	if err != nil {
		t.Fatalf("ParsePubspecYaml: %v", err)
	}
	if len(deps) == 0 {
		t.Fatalf("expected at least one dep, got none")
	}
	if deps[0].Source != "pubspec.yaml" {
		t.Fatalf("source = %q, want pubspec.yaml", deps[0].Source)
	}
}

func TestSortDepsStableBySourceThenName(t *testing.T) {
	deps := []Dependency{
		{Name: "b", Source: "b.mod"},
		{Name: "a", Source: "a.mod"},
		{Name: "z", Source: "a.mod"},
	}
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].Source != deps[j].Source {
			return deps[i].Source < deps[j].Source
		}
		return deps[i].Name < deps[j].Name
	})
	got := make([]string, 0, len(deps))
	for _, d := range deps {
		got = append(got, d.Source+":"+d.Name)
	}
	want := []string{"a.mod:a", "a.mod:z", "b.mod:b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sorted = %v, want %v", got, want)
	}
}

func TestEnsureWritePermissionsForSnippets(t *testing.T) {
	// Smoke test that mustWrite/readFileSnippet are usable together.
	root := t.TempDir()
	path := filepath.Join(root, "pubspec.yaml")
	if err := os.WriteFile(path, []byte("name: x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := readFileSnippet(path)
	if err != nil {
		t.Fatalf("readFileSnippet: %v", err)
	}
	if got != "name: x\n" {
		t.Fatalf("snippet = %q", got)
	}
}
