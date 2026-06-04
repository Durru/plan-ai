# Project Structure

```
plan-ai/
├── cmd/
│   ├── plan-ai/                     # CLI entry point
│   │   └── main.go                  #   Cobra root, all commands
│   └── mcp-server/                  # MCP server entry point
│       └── main.go                  #   stdio JSON-RPC server
│
├── internal/
│   ├── agent/                       # Agent system
│   │   └── agent.go                 #   Intent detection, routing, delegation
│   ├── capabilities/                # Capability registry
│   │   └── capabilities.go          #   Capability definitions
│   ├── change/                      # Change detection engine
│   │   └── change.go                #   Impact analysis, snapshots
│   ├── config/                      # Configuration
│   │   └── config.go                #   Global/project config paths
│   ├── context/                     # Approved context + delivery
│   │   ├── context.go               #   Approved context CRUD + L0-L4 delivery
│   │   └── context_test.go
│   ├── continuous/                  # Continuous planning engine
│   │   ├── continuous.go            #   Event detection, proposal generation
│   │   └── continuous_test.go
│   ├── core/                        # Core metadata
│   │   └── core.go                  #   App metadata, version
│   ├── domain/                      # Canonical entity model
│   │   └── domain.go                #   All domain types
│   ├── ingestion/                   # Input ingestion
│   │   └── ingestion.go             #   Classification and ingestion
│   ├── integrations/                # Integration surface
│   │   └── integrations.go          #   Integration helpers
│   ├── knowledge/                   # Knowledge base
│   │   └── knowledge.go             #   Reusable technical knowledge
│   ├── mcp/                         # MCP tool definitions
│   │   └── tools.go                 #   Tool handler implementations
│   ├── modelstrategy/               # LLM provider registry
│   │   └── modelstrategy.go         #   Provider and budget tracking
│   ├── opencode/                    # OpenCode integration
│   │   └── setup.go                 #   Artifact generation
│   ├── orchestrator/                # Job queue and orchestration
│   │   └── orchestrator.go          #   Async orchestration
│   ├── planning/                    # Planning engine
│   │   ├── planning.go              #   Master/specific plan generation
│   │   └── planning_test.go
│   ├── research/                    # Research engine
│   │   └── research.go              #   Research entries, findings, sources, conclusions
│   ├── scanner/                     # Project scanner
│   │   └── scanner.go               #   Deterministic stack/dependency detection
│   ├── skills/                      # Skills/resources
│   │   └── skills.go                #   Skill definitions
│   ├── store/                       # SQLite persistence layer
│   │   ├── store.go                 #   Store initialization, migrations
│   │   ├── store_test.go
│   │   ├── migrations.go            #   Schema migrations
│   │   ├── repositories.go          #   Repository implementations
│   │   └── repositories_test.go
│   ├── validation/                  # Validation resources
│   │   └── validation.go            #   Validation utilities
│   ├── vision/                      # Vision engine
│   │   └── vision.go                #   Vision draft creation, discovery sessions
│   └── workflows/                   # Workflow execution
│       └── workflows.go             #   Workflow registry
│
├── scripts/
│   └── test-sandbox.sh              # Sandbox validation script
│
├── docs/
│   ├── architecture.md              # Architecture overview
│   ├── installation.md              # Installation guide
│   ├── cli-reference.md             # CLI command reference
│   ├── mcp-reference.md             # MCP server tool reference
│   ├── opencode-integration-guide.md # OpenCode integration guide
│   └── project-structure.md         # This file
│
├── RELEASE_NOTES.md                 # Release notes
├── MVP_REPORT.md                    # MVP completion report
├── TECHNICAL_AUDIT.md               # Technical audit report
├── FEATURE_MATRIX.md                # Feature completeness matrix
├── FINAL_AUDIT_REPORT.md            # Final comprehensive audit
├── README.md                        # Project overview
├── go.mod
├── go.sum
└── LICENSE
```
