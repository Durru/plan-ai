# Quickstart

This guide starts from a fresh checkout and creates a local Plan-AI project store.

```bash
git clone https://github.com/Durru/plan-ai.git
cd plan-ai
bash scripts/install.sh
plan-ai doctor
plan-ai init
plan-ai status
```

## Create planning context

```bash
plan-ai ingest --type prompt --content "Build a planning assistant for product teams. It must use SQLite."
plan-ai vision draft
plan-ai approved add --type requirement "Plans must be stored locally."
plan-ai approved add --type decision "Use SQLite as the source of truth."
plan-ai plan master
plan-ai context
```

## Use V3 product intent commands

Discover an idea:

```bash
plan-ai intent discover "Quiero crear un CRM para talleres mecánicos"
```

Create a durable Product Intent:

```bash
plan-ai intent create \
  --description "CRM for mechanic workshops" \
  --expected-outcome "Workshops can track customers, vehicles, jobs, and follow-ups" \
  --desired-experience "Simple, fast, Spanish-first workflow" \
  --desired-result "A clear implementation plan before coding" \
  --success-definition "A workshop owner can manage jobs without spreadsheets" \
  --failure-definition "The tool becomes a generic CRM with no workshop-specific workflow"
```

Then list, submit, approve, and inspect:

```bash
plan-ai intent list
plan-ai intent submit <pintent_id>
plan-ai intent approve <pintent_id>
plan-ai intent show <pintent_id>
```

## Reduce ambiguity before implementation

```bash
plan-ai discovery init --intent <pintent_id>
plan-ai discovery next --intent <pintent_id>
plan-ai ambiguity analyze --intent <pintent_id>
plan-ai confidence evaluate --intent <pintent_id>
plan-ai alignment framework --intent <pintent_id>
```

## Safety check

Run the sandbox validation before trusting changes:

```bash
bash scripts/test-sandbox.sh
```

The sandbox must end with `SANDBOX_CLEANED`.
