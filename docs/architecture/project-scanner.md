# Project Scanner Architecture

## Purpose

The project scanner is Plan-AI's first deterministic understanding layer for the local repository where it is installed. It produces a reproducible snapshot of repository metadata without using AI, network calls, external planning, or integration-specific behavior.

This scan gives later phases a stable substrate before they attempt planning, research, context selection, or knowledge reuse.

## What It Detects

The scanner detects:

- Git presence and current branch when available.
- Languages by file extension.
- Package managers through manifests and lockfiles.
- Frameworks through deterministic file and dependency evidence.
- Dependencies from common manifests such as `go.mod`, `package.json`, `Cargo.toml`, `requirements.txt`, `pyproject.toml`, `composer.json`, and `pubspec.yaml`.
- Important files and source files, indexed only by path, kind, and size.

## What It Does Not Detect

This phase intentionally does not detect or run:

- Skill Intelligence or skill scanning.
- Planner behavior or automatic plan generation.
- AI-generated analysis.
- Real research workflows.
- MCP behavior.
- OpenCode integration.
- Engram integration.
- Agents or sub-agents.

The scanner only understands local repository evidence with deterministic rules.

## Ignored Directories

The scanner never descends into:

- `.git/`
- `.plan-ai/`
- `node_modules/`
- `vendor/`
- `dist/`
- `build/`
- `coverage/`
- `.tmp/`
- `tmp/`
- `.cache/`

Files larger than 1 MiB are skipped. This keeps scans fast, avoids indexing generated artifacts, and prevents binary or build output from polluting the project model.

## Fingerprint

Each scan stores a lightweight fingerprint built from:

- the absolute project root;
- sorted relevant paths;
- file sizes;
- package-manager evidence files and their modification times.

The fingerprint is not a security hash. Its purpose is to answer: “did this project change in a way that likely matters to Plan-AI?” Future phases can compare fingerprints to decide whether cached context, research, or plan assumptions need refreshing.

## Storage

Scan data is stored in `project.db` only:

- `project_scans` stores the top-level scan row.
- `project_scan_languages` stores detected language counts.
- `project_scan_frameworks` stores framework names and evidence.
- `project_scan_package_managers` stores package manager evidence.
- `project_scan_dependencies` stores parsed dependency names, versions, and sources.
- `project_scan_files` stores indexed file paths, kinds, and sizes.

No file contents are stored in this phase.

## Future Use

Later Plan-AI phases can use scan data for:

- Planner: choose plan templates and implementation constraints from detected stack.
- Research: avoid researching already-detected frameworks from scratch.
- Knowledge: seed reusable project knowledge with stack facts.
- Context Engine: estimate context size from file count, file kinds, and dependency evidence.

The scanner remains deterministic so later AI-assisted layers have a trustworthy baseline.