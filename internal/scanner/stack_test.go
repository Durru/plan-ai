package scanner

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestDetectLanguagesReturnsSortedStats(t *testing.T) {
	files := []ScannedFile{
		{Path: "a.go", Kind: FileKindSource, Size: 10},
		{Path: "b.go", Kind: FileKindSource, Size: 10},
		{Path: "c.ts", Kind: FileKindSource, Size: 10},
		{Path: "README.md", Kind: FileKindDoc, Size: 10},
		{Path: "logo.png", Kind: FileKindOther, Size: 10},
	}

	got := DetectLanguages(files)

	want := []LanguageStats{
		{Language: "Go", Files: 2},
		{Language: "Markdown", Files: 1},
		{Language: "TypeScript", Files: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DetectLanguages = %+v, want %+v", got, want)
	}
}

func TestDetectLanguagesEmptyInput(t *testing.T) {
	got := DetectLanguages(nil)
	if len(got) != 0 {
		t.Fatalf("DetectLanguages(nil) = %+v, want empty", got)
	}
}

func TestDetectPackageManagers(t *testing.T) {
	cases := map[string]struct {
		files map[string]bool
		want  []string
	}{
		"go and pip": {
			files: map[string]bool{"go.mod": true, "requirements.txt": true},
			want:  []string{"go", "pip"},
		},
		"js ecosystem": {
			files: map[string]bool{"package.json": true, "pnpm-lock.yaml": true},
			want:  []string{"pnpm"},
		},
		"nothing": {
			files: map[string]bool{},
			want:  nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := DetectPackageManagers(tc.files)
			names := make([]string, 0, len(got))
			for _, pm := range got {
				names = append(names, pm.Name)
			}
			if len(names) == 0 {
				names = nil
			}
			sort.Strings(names)
			sort.Strings(tc.want)
			if !reflect.DeepEqual(names, tc.want) {
				t.Fatalf("names = %v, want %v", names, tc.want)
			}
		})
	}
}

func TestDetectFrameworksDetectsKnown(t *testing.T) {
	root := t.TempDir()

	// Minimal Next.js + React + Tailwind via package.json.
	mustWrite(t, filepath.Join(root, "package.json"), `{
  "name": "demo",
  "dependencies": {
    "next": "14.2.5",
    "react": "18.3.1"
  },
  "devDependencies": {
    "tailwindcss": "3.4.0"
  }
}`)
	mustWrite(t, filepath.Join(root, "next.config.js"), "module.exports = {}\n")
	mustWrite(t, filepath.Join(root, "tailwind.config.js"), "module.exports = {}\n")

	// Minimal Go + Cobra + SQLite via go.mod.
	mustWrite(t, filepath.Join(root, "go.mod"), `module demo

go 1.25

require (
	github.com/spf13/cobra v1.9.1
	modernc.org/sqlite v1.51.0
)
`)

	scan, err := Default().Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	frameworks := map[string]string{}
	for _, f := range scan.Frameworks {
		frameworks[f.Name] = f.Evidence
	}

	for _, name := range []string{"Next.js", "React", "Tailwind", "Go", "Cobra", "SQLite"} {
		if _, ok := frameworks[name]; !ok {
			t.Fatalf("framework %q missing; got %v", name, frameworks)
		}
	}
}

func TestDetectFrameworksIgnoresUnrelatedProjects(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "main.go"), "package main\n")

	scan, err := Default().Scan(root)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	names := make([]string, 0, len(scan.Frameworks))
	for _, f := range scan.Frameworks {
		names = append(names, f.Name)
	}
	if len(names) != 0 {
		t.Fatalf("expected no frameworks for a plain main.go project, got %v", names)
	}
}
