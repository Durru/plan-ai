package scanner

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectGit inspects root and reports whether it is inside a git working
// tree, returning the current branch when available. The function never
// returns an error; it is intentionally permissive because scans should
// succeed even on bare or partial projects.
func DetectGit(root string) (detected bool, branch string) {
	if info, err := os.Stat(filepath.Join(root, ".git")); err == nil {
		detected = true
		if info.IsDir() {
			if b, ok := readBranchFromHead(root); ok {
				branch = b
				return
			}
		}
		// Fall through to git binary: the .git entry may be a file
		// pointing at a worktree.
	} else if !errors.Is(err, fs.ErrNotExist) {
		// Some other I/O error; report as not detected but keep going.
	}

	if !detected {
		if err := runGitInDir(root, "rev-parse", "--is-inside-work-tree"); err != nil {
			return false, ""
		}
		detected = true
	}

	if branch == "" {
		if out, err := runGitInDirOutput(root, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
			branch = strings.TrimSpace(out)
		}
	}
	if branch == "HEAD" {
		branch = ""
	}
	return detected, branch
}

func readBranchFromHead(root string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(root, ".git", "HEAD"))
	if err != nil {
		return "", false
	}
	first := strings.TrimSpace(strings.SplitN(string(data), "\n", 2)[0])
	const prefix = "ref: refs/heads/"
	if !strings.HasPrefix(first, prefix) {
		return "", false
	}
	return strings.TrimPrefix(first, prefix), true
}

func runGitInDir(root string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	return cmd.Run()
}

func runGitInDirOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
