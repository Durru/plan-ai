package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestKindForPathClassifiesExtensionsAndTests(t *testing.T) {
	cases := []struct {
		path string
		want FileKind
	}{
		{path: "main.go", want: FileKindSource},
		{path: "src/index.ts", want: FileKindSource},
		{path: "app.tsx", want: FileKindSource},
		{path: "scripts/build.js", want: FileKindSource},
		{path: "lib/util.py", want: FileKindSource},
		{path: "Cargo.toml", want: FileKindConfig},
		{path: "package.json", want: FileKindConfig},
		{path: "config.yaml", want: FileKindConfig},
		{path: "go.mod", want: FileKindConfig},
		{path: "go.sum", want: FileKindLock},
		{path: "package-lock.json", want: FileKindLock},
		{path: "yarn.lock", want: FileKindLock},
		{path: "pnpm-lock.yaml", want: FileKindLock},
		{path: "bun.lockb", want: FileKindLock},
		{path: "README.md", want: FileKindDoc},
		{path: "docs/notes.md", want: FileKindDoc},
		{path: "foo_test.go", want: FileKindTest},
		{path: "src/foo_test.py", want: FileKindTest},
		{path: "src/foo.test.ts", want: FileKindTest},
		{path: "src/foo.spec.tsx", want: FileKindTest},
		{path: "src/foo.test.jsx", want: FileKindTest},
		{path: "logo.png", want: FileKindOther},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			got := KindForPath(tc.path)
			if got != tc.want {
				t.Fatalf("KindForPath(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestWalkProjectIgnoresNoiseDirectories(t *testing.T) {
	root := t.TempDir()

	mustWrite(t, filepath.Join(root, "main.go"), "package main\n")
	mustWrite(t, filepath.Join(root, "README.md"), "# Title\n")
	mustMkdir(t, filepath.Join(root, "node_modules"))
	mustWrite(t, filepath.Join(root, "node_modules", "foo.js"), "module.exports = 1\n")
	mustMkdir(t, filepath.Join(root, ".plan-ai"))
	mustWrite(t, filepath.Join(root, ".plan-ai", "bar.go"), "package planai\n")
	mustMkdir(t, filepath.Join(root, "vendor"))
	mustWrite(t, filepath.Join(root, "vendor", "lib.go"), "package vendor\n")
	mustMkdir(t, filepath.Join(root, "dist"))
	mustWrite(t, filepath.Join(root, "dist", "bundle.js"), "x")
	mustMkdir(t, filepath.Join(root, "build"))
	mustWrite(t, filepath.Join(root, "build", "x.txt"), "x")
	mustMkdir(t, filepath.Join(root, "coverage"))
	mustWrite(t, filepath.Join(root, "coverage", "out.html"), "<html/>")
	mustMkdir(t, filepath.Join(root, ".tmp"))
	mustWrite(t, filepath.Join(root, ".tmp", "scratch"), "scratch")
	mustMkdir(t, filepath.Join(root, "tmp"))
	mustWrite(t, filepath.Join(root, "tmp", "scratch"), "scratch")
	mustMkdir(t, filepath.Join(root, ".cache"))
	mustWrite(t, filepath.Join(root, ".cache", "x"), "x")

	files, err := WalkProject(root)
	if err != nil {
		t.Fatalf("WalkProject: %v", err)
	}

	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	sort.Strings(paths)

	want := []string{"README.md", "main.go"}
	if len(paths) != len(want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
	for i, p := range paths {
		if p != want[i] {
			t.Fatalf("paths[%d] = %q, want %q (full=%v)", i, p, want[i], paths)
		}
	}
	for _, p := range paths {
		if filepath.Separator == '\\' && (p != filepath.ToSlash(p)) {
			t.Fatalf("path %q should use forward slashes", p)
		}
	}
}

func TestWalkProjectSkipsLargeFiles(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "small.go"), "package x\n")

	largePath := filepath.Join(root, "huge.go")
	f, err := os.Create(largePath)
	if err != nil {
		t.Fatalf("create huge: %v", err)
	}
	if err := f.Truncate(int64(DefaultMaxFileSize) + 1); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	_ = f.Close()

	files, err := WalkProject(root)
	if err != nil {
		t.Fatalf("WalkProject: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1 (small.go only); got %+v", len(files), files)
	}
	if files[0].Path != "small.go" {
		t.Fatalf("unexpected surviving file: %+v", files[0])
	}
}

func TestWalkProjectReturnsForwardSlashRelativePaths(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "src", "pkg"))
	mustWrite(t, filepath.Join(root, "src", "pkg", "main.go"), "package pkg\n")

	files, err := WalkProject(root)
	if err != nil {
		t.Fatalf("WalkProject: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}
	if files[0].Path != "src/pkg/main.go" {
		t.Fatalf("path = %q, want src/pkg/main.go", files[0].Path)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}
