# Security Policy

## Supported versions

Plan-AI is currently pre-1.0/open-source release candidate software. Security fixes target the `main` branch unless a release branch is created.

## Reporting a vulnerability

Please do not publish exploitable details in a public issue. Report privately to the project maintainer first. If no private channel is available yet, open a minimal GitHub issue that says a security report is available and avoid including secrets, tokens, payloads, or exploit steps.

Include:

- Affected command or package.
- Reproduction steps using a sandbox path.
- Expected vs actual behavior.
- Potential impact.

## Security expectations

Plan-AI must not:

- Commit local project data from `.plan-ai/`.
- Commit SQLite databases, logs, `.env` files, tokens, or generated binaries.
- Mutate real OpenCode configuration unless the user explicitly opts in.
- Delete user data during install, uninstall, or validation without confirmation.

Use `scripts/test-sandbox.sh` and `scripts/test-vps-clean.sh` to validate safety-sensitive changes.
