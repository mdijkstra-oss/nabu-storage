Read `internal/domain/projections/project-entity/view.go` and all referenced entity types.

Generate a fully inlined JSON Schema (Draft 7) for `ProjectArrayView` with:
- All nested types inlined (no $ref usage)
- All struct fields mapped via JSON tags
- Type constraints (maxLength, minLength, enum, format)
- Required fields based on struct definitions

Write to `projections.schemas.json` as a single flat schema object.

LLM-optimized: flat structure for SQL query generation.
