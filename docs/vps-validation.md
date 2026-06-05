# VPS Validation

This guide validates Plan-AI on a clean Linux VPS or in an isolated VPS-style temporary environment.

## Clean VPS checklist

On Ubuntu 22.04/24.04:

```bash
sudo apt-get update
sudo apt-get install -y git golang-go ca-certificates
git clone https://github.com/Durru/plan-ai.git
cd plan-ai
bash scripts/install.sh
plan-ai doctor
plan-ai init
plan-ai status
bash scripts/test-sandbox.sh
bash scripts/test-vps-clean.sh
bash scripts/release-check.sh
```

## What `test-vps-clean.sh` checks

The script simulates a clean user account without touching real user data:

- Creates a temporary HOME.
- Creates a temporary install prefix.
- Builds and installs Plan-AI from the current checkout.
- Runs `plan-ai install`, `doctor`, `init`, `status`.
- Exercises V3 intent, ambiguity, confidence, and alignment commands.
- Runs uninstall while preserving data by default.
- Confirms no real `/root/.plan-ai` or repository `.plan-ai` is created.
- Cleans its temporary directory unless `--keep` is passed.

## Root vs non-root

- Root install default prefix: `/usr/local`.
- Non-root install default prefix: `$HOME/.local`.
- CI/sandbox prefix: `$PLAN_AI_INSTALL_PREFIX`.

## Data safety

Plan-AI runtime data lives in:

- Global: `~/.plan-ai/global.db`
- Project: `<project>/.plan-ai/project.db`

Install never deletes those stores. Uninstall preserves them unless `--purge-data` is explicitly requested and confirmed.

## Failure handling

If validation fails:

1. Re-run with `--keep` where supported.
2. Inspect the temporary directory printed by the script.
3. Run `plan-ai doctor` with the same `HOME`, `PLAN_AI_HOME`, and `PLAN_AI_PROJECT_ROOT` environment variables.
