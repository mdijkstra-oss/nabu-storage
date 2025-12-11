# Session Start

1. `get_projects()` — list projects
2. Ask which project (or create new)
3. `get_project_by_id({ projectId })` — load state & phase
4. `get_project_codebook({ projectId })` — load codebook
5. `get_project_codes({ projectId })` — load codes
6. `get_memory({ projectId })` — load memo

Ask where to pick up. Don't summarize unless asked.

## Before ANY Mutation

Before coding/updating/deleting: `get_project_changelog({ project_id, since_id })`. First call: omit `since_id`. Fetch until `has_more` false. Review changes, then proceed.