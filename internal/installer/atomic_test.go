package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileAtomically_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	err := writeFileAtomically(path, []byte(`{"ok":true}`), 0644)
	if err != nil {
		t.Fatalf("writeFileAtomically: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.TrimSpace(string(data)) != `{"ok":true}` {
		t.Fatalf("content = %q", string(data))
	}
}

func TestWriteFileAtomically_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	// Write first version
	if err := writeFileAtomically(path, []byte(`{"v":1}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Overwrite
	if err := writeFileAtomically(path, []byte(`{"v":2}`), 0644); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"v":2`) {
		t.Fatalf("expected v2, got %s", data)
	}
}

func TestWriteFileAtomically_NoTempLeakOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clean.json")

	if err := writeFileAtomically(path, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// No .tmp files should remain
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Fatalf("temp file leaked: %s", e.Name())
		}
	}
}

func TestWriteFileAtomically_NestedDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "nested.json")

	if err := writeFileAtomically(path, []byte(`{"nested":true}`), 0644); err != nil {
		t.Fatalf("writeFileAtomically nested: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("nested file missing: %v", err)
	}
}

func TestWriteFileAtomically_Permissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "perm.json")

	if err := writeFileAtomically(path, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0777 != 0600 {
		t.Fatalf("expected 0600, got %o", info.Mode()&0777)
	}
}
