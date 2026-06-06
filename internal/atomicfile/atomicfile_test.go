package atomicfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileCreatesFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.json")

	data := []byte(`{"hello": "world"}`)
	if err := WriteFile(target, data, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("contents = %q, want %q", got, data)
	}
}

func TestWriteFileAtomicReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.json")
	if err := os.WriteFile(target, []byte(`{"old": true}`), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	newData := []byte(`{"new": true}`)
	if err := WriteFile(target, newData, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(newData) {
		t.Errorf("contents = %q, want %q", got, newData)
	}
}

func TestWriteFileAtomicNoLeftoverTemp(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.json")
	if err := WriteFile(target, []byte(`{}`), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

func TestWriteFileWithBackupCreatesBackupWhenExists(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.json")
	original := []byte(`{"original": true}`)
	if err := os.WriteFile(target, original, 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	newData := []byte(`{"new": true}`)
	backupPath, err := WriteFileWithBackup(target, newData, 0644, "pre-mcp-write")
	if err != nil {
		t.Fatalf("WriteFileWithBackup: %v", err)
	}

	if backupPath == "" {
		t.Fatal("expected non-empty backup path")
	}
	if !strings.Contains(backupPath, "pre-mcp-write") {
		t.Errorf("backup path %q missing reason tag", backupPath)
	}
	if !strings.HasSuffix(backupPath, ".bak") {
		t.Errorf("backup path %q missing .bak suffix", backupPath)
	}

	// Backup must contain the original content, not the new content.
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backup) != string(original) {
		t.Errorf("backup contents = %q, want %q", backup, original)
	}

	// Target must contain the new content.
	current, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(current) != string(newData) {
		t.Errorf("target contents = %q, want %q", current, newData)
	}
}

func TestWriteFileWithBackupNoBackupWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.json")

	backupPath, err := WriteFileWithBackup(target, []byte(`{"a": 1}`), 0644, "pre-mcp-write")
	if err != nil {
		t.Fatalf("WriteFileWithBackup: %v", err)
	}
	if backupPath != "" {
		t.Errorf("expected empty backup path when no prior file, got %q", backupPath)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 file (target), got %d: %+v", len(entries), entries)
	}
}
