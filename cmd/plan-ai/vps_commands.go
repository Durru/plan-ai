package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newSetupUpdateVpsCommand() *cobra.Command {
	var (
		host    string
		user    string
		port    int
		keyPath string
		dryRun  bool
	)
	cmd := &cobra.Command{
		Use:   "update-vps",
		Short: "Update Plan-AI on a remote VPS via SSH.",
		Long: `SSH into a remote VPS, pull the latest Plan-AI code from GitHub,
build, and install the plan-ai binary.

Requires ssh on PATH. The VPS must have git and go installed.`,
		Example: `  plan-ai setup update-vps --host example.com
  plan-ai setup update-vps --host example.com --user deploy --port 2222
  plan-ai setup update-vps --host example.com --key ~/.ssh/id_vps`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if host == "" {
				return fmt.Errorf("--host is required")
			}

			// Resolve key path if relative
			if keyPath != "" {
				abs, err := filepath.Abs(keyPath)
				if err != nil {
					return fmt.Errorf("resolve key path: %w", err)
				}
				keyPath = abs
				if _, err := os.Stat(keyPath); err != nil {
					return fmt.Errorf("key file %s: %w", keyPath, err)
				}
			}

			// Check ssh is available
			if _, err := exec.LookPath("ssh"); err != nil {
				return fmt.Errorf("ssh not found on PATH: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Plan-AI VPS update\n")
			fmt.Fprintf(out, "  Host: %s\n", host)
			fmt.Fprintf(out, "  User: %s\n", user)
			fmt.Fprintf(out, "  Port: %d\n", port)
			if keyPath != "" {
				fmt.Fprintf(out, "  Key:  %s\n", keyPath)
			}

			if dryRun {
				fmt.Fprintln(out, "\nDry-run mode. SSH command would be:")
				sshCmd := buildSSHCommand(host, user, port, keyPath, remoteUpdateScript())
				fmt.Fprintf(out, "  %s\n", strings.Join(sshCmd.Args, " "))
				return nil
			}

			sshCmd := buildSSHCommand(host, user, port, keyPath, remoteUpdateScript())
			sshCmd.Stdin = nil // script is embedded in args
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr

			fmt.Fprintln(out, "\nConnecting...")
			if err := sshCmd.Run(); err != nil {
				return fmt.Errorf("ssh update failed: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "VPS hostname or IP address (required)")
	cmd.Flags().StringVar(&user, "user", "root", "SSH user")
	cmd.Flags().IntVar(&port, "port", 22, "SSH port")
	cmd.Flags().StringVar(&keyPath, "key", "", "SSH private key path")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print SSH command without executing")
	return cmd
}

// buildSSHCommand returns an exec.Cmd that runs the remote script on the VPS.
func buildSSHCommand(host, user string, port int, keyPath, remoteScript string) *exec.Cmd {
	args := []string{
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "LogLevel=ERROR",
		"-p", strconv.Itoa(port),
	}
	if keyPath != "" {
		args = append(args, "-i", keyPath)
	}
	args = append(args, fmt.Sprintf("%s@%s", user, host), "bash", "-s")

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = strings.NewReader(remoteScript)
	return cmd
}

// newUpdateVpsCommand returns a top-level "plan-ai update-vps" command
// that delegates to the setup subcommand implementation.
func newUpdateVpsCommand() *cobra.Command {
	base := newSetupUpdateVpsCommand()
	cmd := &cobra.Command{
		Use:   "update-vps",
		Short: "Update Plan-AI on a remote VPS via SSH.",
		Long: `SSH into a remote VPS, pull the latest Plan-AI code from GitHub,
build, and install the plan-ai binary.

Requires ssh on PATH. The VPS must have git and go installed.`,
		Example: `  plan-ai update-vps --host example.com
  plan-ai update-vps --host example.com --user deploy --port 2222
  plan-ai update-vps --host example.com --key ~/.ssh/id_vps`,
		RunE: base.RunE,
	}
	cmd.Flags().AddFlagSet(base.Flags())
	return cmd
}

// remoteUpdateScript returns the bash script executed on the remote VPS.
func remoteUpdateScript() string {
	return `set -euo pipefail

PLAN_AI_DIR="${PLAN_AI_DIR:-${HOME}/plan-ai}"
BIN_DIR="${BIN_DIR:-${HOME}/.local/bin}"

if [[ -d "${PLAN_AI_DIR}/.git" ]]; then
  printf 'Pulling latest in %s...\n' "${PLAN_AI_DIR}"
  cd "${PLAN_AI_DIR}"
  git pull --ff-only
else
  printf 'Cloning Plan-AI to %s...\n' "${PLAN_AI_DIR}"
  mkdir -p "$(dirname "${PLAN_AI_DIR}")"
  git clone https://github.com/Durru/plan-ai.git "${PLAN_AI_DIR}"
  cd "${PLAN_AI_DIR}"
fi

cd "${PLAN_AI_DIR}"

printf 'Building...\n'
go build -o "${BIN_DIR}/plan-ai" ./cmd/plan-ai
chmod 0755 "${BIN_DIR}/plan-ai"

INSTALLED="$("${BIN_DIR}/plan-ai" --version 2>/dev/null || "${BIN_DIR}/plan-ai" version 2>/dev/null || true)"
printf 'Installed: %s/plan-ai\n' "${BIN_DIR}"
printf 'Version: %s\n' "${INSTALLED:-plan-ai}"

"${BIN_DIR}/plan-ai" install 2>/dev/null || true

printf 'UPDATE_VPS_OK\n'
`
}
