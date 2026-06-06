# ADR 0021: Safe OpenCode Auto-Configuration

**Status:** Accepted
**Date:** 2026-06-06
**Phase:** 3
**Supersedes:** ADR 0017 (Read-Only Detection)

## Context

ADR 0017 (2026-06-03) mandated that Plan-AI detect OpenCode configuration
but never write to it. The rationale was "safe coexistence" — Plan-AI could
not corrupt a user's OpenCode setup.

Phase 1 introduced external project storage (`~/.plan-ai/projects/<slug>/`)
as the default. Phase 2 introduced the install-once lifecycle:
`plan-ai install`, `plan-ai update`, `plan-ai uninstall`, `plan-ai doctor`.
Phase 2 also began writing the `mcpServers.plan-ai` entry to
`<opencode-config-dir>/config.json` via `opencode.SetupMCPConfig`
(`internal/opencode/setup.go:528`) so that an installed Plan-AI can be
invoked by OpenCode without a separate manual step.

This creates a conflict: ADR 0017 says "never write to OpenCode", but the
implementation writes `mcpServers.plan-ai` whenever the user runs
`plan-ai install`. The product vision (`docs/architecture/plan-ai-product-vision.md`)
explicitly requires install/setup to auto-register the MCP server and
configure integrations, because the alternative — telling every user to
hand-edit `config.json` — defeats the purpose of `install` being a single
command.

The question is no longer "should Plan-AI write OpenCode config?" — that
ship has sailed. The question is: **under what conditions is writing
OpenCode config safe, and what guarantees must the implementation give?**

## Decision

Plan-AI **may** write to OpenCode configuration during install, update,
uninstall, and doctor-driven repair. Writes are allowed **only** when all
of the following hold:

1. The user has run `plan-ai install` (or `plan-ai update`) with intent
   to integrate, and the install path is not a sandbox.
2. The target directory is either (a) `$OPENCODE_CONFIG_DIR` if set, or
   (b) `<homeRoot>/.config/opencode` where `homeRoot` is the
   OS-level real home of the current user (resolved via `getpwuid_r`,
   **not** `os.UserHomeDir()` which respects `$HOME`).
3. The target is **not** the OS-level real home of the current user
   **unless** `--allow-real-opencode` was passed. This is the
   refuse-to-touch-real-OpenCode-config guard.
4. The write is atomic (write-temp + rename), preserves all non-Plan-AI
   keys, and produces a backup when the write would otherwise lose data.
5. The write sets ownership markers so a future `plan-ai uninstall` or
   `plan-ai doctor` can identify Plan-AI-owned artifacts.

The previous "read-only detection" stance in ADR 0017 is **superseded**.
Detection is still performed (Phase 1 + Phase 2 still ship the
`opencode.Detector`), but it is now a step in the install pipeline, not
the entirety of the integration.

## Rationale

The product is no longer viable without auto-configuration. A user who
runs `plan-ai install` expects the tool to work end-to-end with OpenCode;
asking them to hand-edit `config.json` after install breaks the "install
once" contract (Phase 2) and creates support burden.

At the same time, blind writes to `~/.config/opencode` would corrupt
user setups if Plan-AI's heuristics were wrong about which keys to
preserve, which schema version OpenCode uses, or which `mcp` key shape
(legacy `"mcp"` map vs new `"mcpServers"` map) the user's config follows.

The compromise: writes are allowed, but gated by a small set of rules
that make the failure modes recoverable.

### Why gate on `--allow-real-opencode`

`--allow-real-opencode` is the user's explicit consent that "yes, write
to my real `~/.config/opencode`". Without it, the installer refuses and
errors out (or, in `plan-ai doctor`, reports a warning). This is the same
pattern as `cargo install --path` requiring user intent before mutating
shell profiles, and matches the safety precedent in the installer's
existing refuse-to-write-real-OpenCode check
(`internal/installer/sync.go:syncOpenCodeConfig`).

### Why gate on `getpwuid_r`, not `$HOME`

`os.UserHomeDir()` reads `$HOME`. In a test, `$HOME` is sandboxed, so
`os.UserHomeDir()` returns the sandbox path — and a "refuse to write
real OpenCode" check that compares against it is a no-op. `os/user`'s
`Current().HomeDir` reads the passwd database via `getpwuid_r` and is
not affected by the environment. This is the same reason `os.UserCacheDir`
is preferred in security-sensitive code.

### Why preserve all non-Plan-AI keys

OpenCode configs carry user state: providers, theme, custom agents,
keybindings, custom commands. Plan-AI only owns the `mcpServers.plan-ai`
(and legacy `mcp.plan-ai`) key. The merge implementation in
`internal/installer/sync.go:generateOpenCodeConfigContent` reads the
existing config, strips a known-invalid set (`invalidOpenCodeKeys`),
adds/updates the Plan-AI entry, and writes the merged result back. A
complete overwrite would be a regression.

### Why write atomically

A non-atomic write that crashes mid-write would leave the user with a
half-written `config.json` and a broken OpenCode. The installer uses
`internal/installer/atomic.go:writeFileAtomically` for the data directory
already; the same helper must be used for any OpenCode config write.

## Mechanics

### Ownership markers

A write to OpenCode config is Plan-AI-owned if and only if it sets the
`mcpServers.plan-ai` key (or the legacy `mcp.plan-ai` key). This is the
single, structural ownership marker. There is no separate
`"owned_by": "plan-ai"` field — the key name itself is the marker, by
convention across both old and new OpenCode config formats.

A `plan_ai` MCP server entry must:

- Have key `"plan-ai"` (or be removed by `plan-ai uninstall`).
- Have `command` whose first element is `config.MCPCommand(binDir)`
  (typically `[<binDir>/plan-ai-mcp-server]`).
- Be the only `plan-ai`-prefixed key in `mcpServers` (and `mcp`) —
  Doctor enforces this with `countPlanAIEntries`.

There is **no** ownership marker for non-MCP files (profiles, agents,
prompts, workflows). Those are written by `internal/opencode/setup.go`
under a `plan-ai-` filename prefix (`plan-ai-workflows.json`, etc.) so
that a future `plan-ai uninstall` knows what to clean up. The
`opencode-sync.json` marker (in `<project>/.plan-ai/`) records what
artifacts exist and when they were last synced.

### Backup and rollback

A write that **modifies** an existing OpenCode config file MUST back up
the original first. The backup naming convention is:

```
<config-path>.<reason>.<UTC-timestamp>
```

where `<reason>` is one of:
- `.invalid.` — original was unparseable JSON/JSONC
- `.stripped.` — original contained keys in `invalidOpenCodeKeys`
  (`providers`, `provider.list`, `app.agents`)

A write that **creates** a new config file does not need a backup (no
prior content to lose). A write that **only updates the `plan-ai`
entry** in an otherwise-unchanged config does not need a backup (atomic
write + the prior content is recoverable from the OS file system if
needed).

Backup files are left in place. `plan-ai doctor` does not auto-delete
them; it warns if more than 10 backup files exist in the OpenCode
config dir. The user is responsible for cleanup; auto-deletion would
defeat the purpose of a backup.

A rollback is "restore the most recent `.invalid.` or `.stripped.`
backup". This is a manual operation, not a `plan-ai` command, because
Plan-AI is conservative about modifying OpenCode config outside the
install/update/uninstall lifecycle.

### Consent rules

| Command                          | Default behavior                                  | Override flag                |
|----------------------------------|---------------------------------------------------|------------------------------|
| `plan-ai install`                | Refuse to write real `~/.config/opencode`         | `--allow-real-opencode`      |
| `plan-ai install` w/ `$OPENCODE_CONFIG_DIR` | Write to `$OPENCODE_CONFIG_DIR` if it equals the OS-real home **and** user is running interactively (TTY) | `--allow-real-opencode` for non-TTY/CI |
| `plan-ai update`                 | Same as `install`                                 | `--allow-real-opencode`      |
| `plan-ai uninstall` (full)       | Backs up OpenCode config, strips Plan-AI entry,   | (none — uninstall is non-destructive by design) |
|                                  | removes data dir. The backup is preserved.         |                              |
| `plan-ai doctor`                 | **Never** writes; only reports.                    | (none)                       |
| `plan-ai doctor --fix`           | Repairs stale state (re-runs `inst.Sync` for the  | (none — `--fix` itself is the consent) |
|                                  | local installation only) without touching real    |                              |
|                                  | OpenCode config.                                  |                              |

The `--allow-real-opencode` flag must be passed explicitly; it has no
short form and is not implied by any other flag. CI scripts that need
real OpenCode integration must set `--allow-real-opencode` AND verify
the test fixture is in a sandbox, not the developer's real machine.

### `$OPENCODE_CONFIG_DIR` and tests

`$OPENCODE_CONFIG_DIR` takes precedence over `<home>/.config/opencode`
in all installer and setup paths. The test helper
`cmd/plan-ai/main_test.go:executeCommand` defaults
`OPENCODE_CONFIG_DIR` to `<home>/.config/opencode` so the installer's
refuse-to-write-real-OpenCode check (which compares the target against
`os/user.Current().HomeDir`) does not trip in tests.

In test sandboxes:
- `HOME` is set to a `t.TempDir()`.
- `OPENCODE_CONFIG_DIR` defaults to `<HOME>/.config/opencode` (so the
  refuse check sees a directory inside the sandbox).
- The refuse check uses `os/user.Current().HomeDir` (getpwuid_r) which
  ignores `HOME` — this is correct: the installer's idea of "real"
  is the OS real home, not the env-tampered one.

The refuse check must use `os/user.Current().HomeDir`, not
`os.UserHomeDir()`. The latter respects `$HOME` and would make the
check always compare `sandbox` to `sandbox` and pass — defeating the
guard. This is the same lesson as the existing fix in
`internal/installer/installer.go:openCodeConfigDir(homeOverride)`
where the call site now passes `inst.HomeDir` instead of letting
`os.UserHomeDir` decide.

### Doctor issue taxonomy

Doctor reports integration health via `DoctorReport.Issues`
(`internal/installer/types.go`). Phase 2 added these codes; Phase 3
codifies them:

| Code                              | Severity | Meaning                                                                                  |
|-----------------------------------|----------|------------------------------------------------------------------------------------------|
| `registered_binary_missing`       | warn     | State records MCP binary at `<BinDir>/plan-ai-mcp-server` but it is not on disk. Run `plan-ai update` to repair. |
| `duplicate_opencode_registration` | warn     | OpenCode config has more than one `plan-ai*` MCP entry. Run `plan-ai uninstall` then `plan-ai install`. |
| `stale_state_opencode_missing`    | warn     | State exists but OpenCode config has no `plan-ai` entry. Run `plan-ai update` to re-sync. |

Future codes MUST use the same shape: `<domain>_<problem>` in
snake_case, severity one of `info | warn | error`. A `DoctorIssue` is
advisory; it never aborts a command. Resolution is always a follow-up
command (`update`, `uninstall`, `install`), not an automatic fix
unless the user passes `--fix`.

## Consequences

### Positive

- `plan-ai install` does what users expect (Phase 2 contract preserved).
- The product can ship end-to-end OpenCode integration without
  requiring users to hand-edit `config.json`.
- All writes are recoverable: backups are preserved; ownership is
  structural; uninstall is surgical (Phase 2).
- Tests work because the refuse check uses `getpwuid_r`, not `$HOME`.

### Negative

- Plan-AI now shares filesystem with OpenCode. A future bug in the
  merge logic could corrupt user config. Mitigated by backups and
  atomic writes.
- Two MCP key shapes (`mcp` legacy, `mcpServers` new) must both be
  handled. Mitigated by the dual-format reader/writer in
  `internal/installer/sync.go`.
- `--allow-real-opencode` adds a flag users must learn. Acceptable
  because the alternative — auto-detect "real vs sandbox" — is
  fragile (any test runner is a sandbox; any container is a sandbox).

### Neutral

- ADR 0017 is superseded. Anyone relying on the "never write" guarantee
  must read this ADR. The `docs/opencode-integration.md` and
  `docs/opencode-integration-guide.md` documents are updated
  accordingly (Phase 2 already updated `docs/opencode-integration.md`
  to mention `--allow-real-opencode`).

## Alternatives Considered

**Read-only by default with manual MCP registration command.** Users
run `plan-ai setup opencode --real` explicitly. Rejected: defeats the
install-once contract and shifts complexity to users.

**No OpenCode auto-configuration; ship a `plan-ai mcp serve` command
that users wire up themselves.** Rejected: same as above; the MCP
config in `config.json` is the standard OpenCode convention and
Plan-AI is opinionated about integration.

**Use a Plan-AI-only sidecar config (e.g., `~/.config/opencode/plan-ai.json`)
that OpenCode auto-loads.** Rejected: OpenCode does not currently
support sidecar config files; relying on undocumented behavior is
worse than explicit `mcpServers.plan-ai` mutation.

**Always write to real `~/.config/opencode` without consent.** Rejected:
silent writes to a directory the user controls is a regression from
the "safe coexistence" intent of ADR 0017 and creates an immediate
trust problem.

## Enforcement

This ADR's rules are enforced by the following code, which Phase 2
landed and Phase 3 hardened:

- `internal/opencode/setup.go:SetupMCPConfig` — single authority for
  writing the `mcpServers.plan-ai` entry; respects `$OPENCODE_CONFIG_DIR`;
  uses `atomicfile.WriteFileWithBackup` to create a `<path>.pre-mcp-write.<UTC>.bak`
  backup when the target `config.json` already exists.
- `internal/atomicfile/atomicfile.go` — crash-safe write+rename primitive
  (`WriteFile`, `WriteFileWithBackup`); the target file is never observed in
  a partial state.
- `internal/installer/installer.go:syncOpenCodeConfig` — refuse-to-write
  check; uses `os/user.Current().HomeDir`; calls `SetupMCPConfig`.
- `internal/installer/installer.go:openCodeConfigDir(homeOverride)` —
  consistent with `inst.HomeDir`; bypasses `os.UserHomeDir()` in tests.
- `internal/installer/installer.go:Doctor` — emits the three issue
  codes above when invariants are violated.
- `internal/installer/sync.go:removePlanAIFromOpenCodeConfig` — strips
  Plan-AI entry from both `mcp` and `mcpServers` for surgical uninstall;
  uses `writeFileAtomically` (now wrapping `atomicfile.WriteFile`).
- `internal/installer/sync.go:generateOpenCodeConfigContent` —
  merge-then-write with `.invalid.` and `.stripped.` backups.
- `cmd/plan-ai/setup_commands.go:newInstallCommand` and
  `cmd/plan-ai/installer_commands.go:newUpdateCommand` — pass
  `--allow-real-opencode` through to the installer.
- `cmd/plan-ai/info_commands.go:newDoctorCommand` — `--fix` flag
  re-runs `inst.Sync` for `stale_state_opencode_missing` and
  `registered_binary_missing`; never writes to real OpenCode config
  (the refuse-check still applies; `AllowReal=false` is hard-coded).

Tests encoding these rules:

- `TestInstallRespectsOpenCodeConfigDir` — `$OPENCODE_CONFIG_DIR`
  wins over `<home>/.config/opencode`.
- `TestInstallDefaultUsesInstallerPath` — install path produces a
  valid state and integration.
- `TestUpdateRefreshesStateAndIntegrations` — update is idempotent.
- `TestUninstallFullRemovesOpenCodeRegistration` — uninstall strips
  the `mcpServers.plan-ai` entry while leaving other entries intact.
- `TestDoctorDetectsMissingRegisteredBinary` — `registered_binary_missing`.
- `TestDoctorDetectsDuplicateOpenCodeRegistration` — `duplicate_opencode_registration`.
- `TestDoctorFixRepairsStaleState` — `doctor --fix` restores vanished
  OpenCode config by re-running installer sync (phase3).
- `TestDoctorFixRepairsRegisteredBinaryMissing` — `doctor --fix` detects
  and handles `registered_binary_missing` (phase3).
- `TestSetupMCPConfig_BackupWhenConfigExists` — `SetupMCPConfig` creates
  a backup when `config.json` already exists (phase3, unit).
- `TestSetupMCPConfig_NoBackupWhenConfigAbsent` — no backup created when
  `config.json` is new (phase3, unit).
- `TestSetupMCPConfig_AtomicNoTempLeftover` — no `.tmp-*` files remain
  after write (phase3, unit).
- `TestSetupMCPConfigIsAtomic` — end-to-end: install → config.json is
  valid JSON, no temp leftovers (phase3, integration).
- `TestWriteFileCreatesFile`, `TestWriteFileAtomicReplacesExisting`,
  `TestWriteFileWithBackupCreatesBackupWhenExists`, etc. — atomicfile
  primitives are correct (phase3, unit).

A future regression in any of these tests is a regression in this ADR.
