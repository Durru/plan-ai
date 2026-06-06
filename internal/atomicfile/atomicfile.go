// Package atomicfile provides crash-safe file writes.
//
// The write+rename strategy guarantees that a target file is never observed
// in a partial state: the data is fully written to a sibling temp file and
// then atomically renamed into place. This is the same primitive Plan-AI
// uses for the OpenCode MCP config (SetupMCPConfig) and for installer
// data-directory writes.
package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WriteFile writes data to path atomically: a temp file is created in the
// same directory, written fully, fsync'd, and renamed into place. The
// target file is never observed in a partial state.
//
// If the target already exists it is replaced; the caller is responsible
// for backing it up first if needed (see WriteFileWithBackup).
func WriteFile(path string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*"+filepath.Ext(path))
	if err != nil {
		return fmt.Errorf("create temp in %s: %w", dir, err)
	}
	tmpName := tmp.Name()

	defer func() {
		tmp.Close()
		if err != nil {
			os.Remove(tmpName)
		}
	}()

	if _, werr := tmp.Write(data); werr != nil {
		return fmt.Errorf("write temp %s: %w", tmpName, werr)
	}
	if cerr := tmp.Chmod(perm); cerr != nil {
		return fmt.Errorf("chmod temp %s: %w", tmpName, cerr)
	}
	if cerr := tmp.Sync(); cerr != nil {
		return fmt.Errorf("fsync temp %s: %w", tmpName, cerr)
	}
	if cerr := tmp.Close(); cerr != nil {
		return fmt.Errorf("close temp %s: %w", tmpName, cerr)
	}
	if rerr := os.Rename(tmpName, path); rerr != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmpName, path, rerr)
	}
	return nil
}

// WriteFileWithBackup backs up the existing file (if any) and then
// atomically writes new data. The backup filename is
//
//	<path>.<reason>.<UTC-timestamp>
//
// where timestamp is RFC3339 with `:` replaced by `-` so it is a valid
// filename on every platform. If no existing file is present, no backup
// is created.
//
// Use this for Plan-AI config writes that may modify a user's existing
// file (e.g. OpenCode's config.json) so the prior content is recoverable
// from the backup if the new write is wrong.
//
// backupReason is a short snake_case tag identifying why the backup was
// created, e.g. "invalid" or "stripped".
func WriteFileWithBackup(path string, data []byte, perm os.FileMode, backupReason string) (backupPath string, err error) {
	if existing, statErr := os.Stat(path); statErr == nil && !existing.IsDir() {
		ts := time.Now().UTC().Format("2006-01-02T15-04-05Z")
		backupPath = path + "." + backupReason + "." + ts + ".bak"
		if copyErr := copyFile(path, backupPath); copyErr != nil {
			return "", fmt.Errorf("backup %s -> %s: %w", path, backupPath, copyErr)
		}
	}
	if werr := WriteFile(path, data, perm); werr != nil {
		return backupPath, werr
	}
	return backupPath, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
