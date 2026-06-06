# MCP Reference

Plan-AI exposes a stdio-based MCP (Model Context Protocol) server through `plan-ai mcp serve`. The server uses JSON-RPC 2.0 over stdin/stdout.

## Starting the server

```bash
go run ./cmd/plan-ai mcp serve
```

Or from a built binary:

```bash
plan-ai mcp serve
```

## Tool categories

- **Project** — initialization, status, scanning (tools 1–5)
- **Plan** — master plan, specific plan, implementation docs (tools 6–10)
- **Research & Knowledge** — research entries, knowledge base (tools 11–18)
- **Context** — approved context management, context overview (tools 19–21)
- **Agent & Continuous** — agent processing, continuous planning (tools 22–25)
- **Change & Export** — change detection, snapshots, export (tools 26–28)

## Tool reference

### Project tools

#### `plan_ai_install`

Install global persistence.

**Input schema:** `{}`

#### `plan_ai_init`

Initialize project store.

**Input schema:** `{}`

#### `plan_ai_status`

Show persistence and domain status.

**Input schema:** `{}`

#### `plan_ai_scan`

Deterministic project scan.

**Input schema:** `{}`

#### `plan_ai_capabilities`

List registered capabilities.

**Input schema:** `{}`

---

### Plan tools

#### `plan_ai_plan_master`

Generate master plan.

**Input schema:** `{}`

#### `plan_ai_plan_specific`

Generate specific plan.

**Input schema:** `{}`

#### `plan_ai_plan_impl_doc`

Generate implementation document.

**Input schema:** `{}`

#### `plan_ai_plan_approve`

Approve the current plan.

**Input schema:** `{}`

#### `plan_ai_plan_list`

List all plans.

**Input schema:** `{}`

---

### Research & Knowledge tools

#### `plan_ai_research_add`

Add a research entry.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "topic":    { "type": "string" },
    "summary":  { "type": "string" },
    "source":   { "type": "string" }
  },
  "required": ["topic", "summary"]
}
```

#### `plan_ai_research_list`

List research entries.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "topic":  { "type": "string" },
    "limit":  { "type": "number", "default": 20 }
  }
}
```

#### `plan_ai_research_get`

Get a research entry by ID.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "id": { "type": "string" }
  },
  "required": ["id"]
}
```

#### `plan_ai_research_findings_add`

Add a finding to a research entry.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "research_id":  { "type": "string" },
    "finding":      { "type": "string" },
    "source":       { "type": "string" }
  },
  "required": ["research_id", "finding"]
}
```

#### `plan_ai_research_sources_add`

Add a source URL to a research entry.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "research_id":  { "type": "string" },
    "url":          { "type": "string" },
    "description":  { "type": "string" }
  },
  "required": ["research_id", "url", "description"]
}
```

#### `plan_ai_research_conclusions_add`

Add a conclusion to a research entry.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "research_id":  { "type": "string" },
    "conclusion":   { "type": "string" }
  },
  "required": ["research_id", "conclusion"]
}
```

#### `plan_ai_knowledge_add`

Add a knowledge entry.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "topic":    { "type": "string" },
    "content":  { "type": "string" },
    "source":   { "type": "string" }
  },
  "required": ["topic", "content"]
}
```

#### `plan_ai_knowledge_list`

List knowledge entries.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "topic":  { "type": "string" },
    "limit":  { "type": "number", "default": 20 }
  }
}
```

#### `plan_ai_knowledge_get`

Get a knowledge entry by ID.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "id": { "type": "string" }
  },
  "required": ["id"]
}
```

---

### Context tools

#### `plan_ai_approved_add`

Add an approved context item.

**Input schema:**
```json
{
  "type": "object",
  "properties": {
    "type":    { "type": "string" },
    "content": { "type": "string" }
  },
  "required": ["type", "content"]
}
```

#### `plan_ai_context_overview`

Executive context overview.

**Input schema:** `{}`

#### `plan_ai_doctor`

Check store and integration health.

**Input schema:** `{}`

---

### Agent & Continuous tools

#### `plan_ai_agent_status`

Show agent system status.

**Input schema:** `{}`

#### `plan_ai_agent_process`

Trigger agent processing.

**Input schema:** `{}`

#### `plan_ai_continuous_status`

Show continuous planning status.

**Input schema:** `{}`

#### `plan_ai_continuous_events`

List continuous planning events.

**Input schema:** `{}`

#### `plan_ai_continuous_proposals`

List continuous planning proposals.

**Input schema:** `{}`

---

### Change & Export tools

#### `plan_ai_change_detect`

Detect codebase changes since last scan.

**Input schema:** `{}`

#### `plan_ai_snapshot_create`

Create a project snapshot.

**Input schema:** `{}`

#### `plan_ai_export_markdown`

Export plans and context as Markdown.

**Input schema:** `{}`
