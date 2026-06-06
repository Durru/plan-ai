#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

printf '==> gofmt check\n'
gofmt -w cmd internal
if [[ -n "$(gofmt -l cmd internal)" ]]; then
  printf 'gofmt still reports unformatted files\n' >&2
  gofmt -l cmd internal >&2
  exit 1
fi

printf '==> go test\n'
go test ./...

printf '==> go vet\n'
go vet ./...

printf '==> go build\n'
go build ./...

printf '==> sandbox validation\n'
bash scripts/test-sandbox.sh

printf '==> VPS-style clean validation\n'
bash scripts/test-vps-clean.sh

printf '==> repository safety checks\n'
test ! -e "$ROOT_DIR/.plan-ai"
test ! -e "/root/.plan-ai"

if [[ -d "$ROOT_DIR/.tmp" ]]; then
  if find "$ROOT_DIR/.tmp" -mindepth 1 -print -quit | grep -q .; then
    printf '.tmp is not empty after release checks\n' >&2
    find "$ROOT_DIR/.tmp" -mindepth 1 -maxdepth 2 >&2
    exit 1
  fi
fi

if git ls-files | grep -E '(^|/)\.plan-ai/|\.db$|\.sqlite3?$|\.env$|\.log$|(^|/)(plan-ai)$' >/dev/null; then
  printf 'tracked runtime artifact detected\n' >&2
  git ls-files | grep -E '(^|/)\.plan-ai/|\.db$|\.sqlite3?$|\.env$|\.log$|(^|/)(plan-ai)$' >&2
  exit 1
fi

for required in README.md LICENSE CONTRIBUTING.md CODE_OF_CONDUCT.md SECURITY.md CHANGELOG.md .editorconfig docs/install.md docs/quickstart.md docs/manual-validation.md docs/vps-validation.md docs/opencode-integration.md docs/troubleshooting.md docs/cli-reference.md scripts/install.sh scripts/uninstall.sh scripts/test-vps-clean.sh; do
  test -s "$required"
done

test -f .github/workflows/ci.yml

git check-ignore -q .plan-ai/test.db
git check-ignore -q sample.db
git check-ignore -q .env
git check-ignore -q plan-ai

if git check-ignore -q cmd/plan-ai/main.go; then
  printf 'cmd/plan-ai/main.go must not be ignored\n' >&2
  exit 1
fi

printf 'RELEASE_CHECK_OK\n'
