# Storage Schema

## Global store

Definitive global tables:

- `global_config`
- `global_tools`
- `global_integrations`
- `global_skills`
- `global_skill_cache`
- `global_knowledge`
- `global_research`
- `global_templates`
- `global_model_profiles`
- `global_logs`

Legacy global tables retained:

- `global_metadata`
- `global_settings`
- `known_projects`

## Project store

Definitive project tables:

- `projects`
- `raw_inputs`
- `ingested_sources`
- `visions`
- `requirements`
- `constraints`
- `decisions`
- `decision_history`
- `research_entries`
- `research_sources`
- `research_findings`
- `knowledge_objects`
- `knowledge_relations`
- `knowledge_references`
- `master_plans`
- `specific_plans`
- `implementation_documents`
- `phases`
- `tasks`
- `task_steps`
- `validations`
- `snapshots`
- `change_requests`
- `impact_reports`
- `context_views`
- `context_chunks`
- `agent_runs`
- `subagent_outputs`
- `project_scans`
- `file_snapshots`
- `file_change_events`

Legacy/additional project tables retained:

- `project_metadata`
- `project_settings`
- `project_state`
- `plans`
- `project_scan_languages`
- `project_scan_frameworks`
- `project_scan_package_managers`
- `project_scan_dependencies`
- `project_scan_files`
- `knowledge_tags`
- `research_conclusions`
- `research_tags`
- `research_knowledge_links`

## Search tables

- `knowledge_objects_fts`
- `research_entries_fts`
- `implementation_documents_fts`
- `raw_inputs_fts`

The FTS tables are preparation for future engines. Current repository search keeps a simple SQL `LIKE` fallback.
