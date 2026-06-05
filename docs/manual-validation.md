# Manual Validation

Use this scenario to validate Plan-AI from a user's point of view.

## Scenario

User request:

> Quiero crear un CRM para talleres mecánicos

Expected behavior: Plan-AI should preserve intent, expose ambiguity, ask for missing information, and produce alignment context before implementation starts.

## Steps

```bash
plan-ai doctor
plan-ai init
plan-ai status
```

### 1. Discover intent

```bash
plan-ai intent discover "Quiero crear un CRM para talleres mecánicos"
```

Expected output includes:

- `Discovery Result:`
- `Detected Intent:`
- `Classification:`
- objectives, gaps, or questions

### 2. Create Product Intent

```bash
plan-ai intent create \
  --description "CRM for mechanic workshops" \
  --expected-outcome "Workshop staff can manage customers, vehicles, repair orders, reminders, and status updates" \
  --desired-experience "Fast daily workflow for non-technical Spanish-speaking users" \
  --desired-result "A validated plan before implementation" \
  --success-definition "A mechanic can find a vehicle, see active work, and notify the customer quickly" \
  --failure-definition "The product becomes a generic CRM that ignores workshop operations"
```

Capture the returned `pintent_...` ID.

### 3. Run progressive discovery

```bash
plan-ai discovery init --intent <pintent_id>
plan-ai discovery next --intent <pintent_id>
plan-ai discovery answer --intent <pintent_id> --question <question_id> --answer "The first version supports customers, vehicles, repair orders, and WhatsApp-ready customer updates."
plan-ai discovery v3-status --intent <pintent_id>
```

Expected output includes question counts and progression hints.

### 4. Analyze ambiguity and confidence

```bash
plan-ai ambiguity analyze --intent <pintent_id>
plan-ai confidence evaluate --intent <pintent_id>
```

Expected output includes an ambiguity score and final intent confidence percentage.

### 5. Review alignment

```bash
plan-ai intent submit <pintent_id>
plan-ai intent approve <pintent_id>
plan-ai alignment context --intent <pintent_id>
plan-ai alignment review --intent <pintent_id> --outcome "Workshop CRM MVP" --plan "Build customers, vehicles, repair orders, and status tracking" --task "Create customer and vehicle schema"
plan-ai alignment framework --intent <pintent_id>
```

Expected behavior: the report should tie outcome, plan, and task back to the approved Product Intent.

## Pass criteria

- Commands exit successfully.
- Product Intent IDs are persisted and can be shown/listed.
- Ambiguity and confidence reports are deterministic.
- Alignment output references the original user intent.
- No `.plan-ai/` data is committed.
