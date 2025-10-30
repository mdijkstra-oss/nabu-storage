# OpenAPI Spec

Generate a complete OpenAPI 3.0 specification for this repository.

## Output Configuration

- **Output file**: `open-api-spec.generated.yml` (in repository root)
- **Name**: hermes-relay-api
- **Version**: 1
- **Servers**: http://localhost:8080
- **Always override existing file** - use Write tool to replace

## System Overview

This is a CQRS (Command Query Responsibility Segregation) system for qualitative coding research. All examples should use realistic qualitative coding contexts (interviews, transcripts, climate research, policy analysis, etc.).

## Key Files to Reference

### Core CQRS Structure
- **Message type**: `internal/cqrs/message.go` - defines `Message[T]` and `AnyMessage` structure
  - All commands/events follow this structure
  - Key fields: `action`, `type`, `aggregateId`, `aggregateType`, `payload`, `timestamp`

### HTTP Handlers & Routes
- **Routes definition**: `internal/handlers/routes.go` - contains ALL endpoint definitions
- **Handler logic**: `internal/handlers/http/handlers.go` - command/query processing
- **Error handling**: `internal/handlers/http/processors.go` - error response format

### Entities (Commands & Payloads)
For each entity in `internal/domain/entities/{entity}/`:
- **commands.go** - command action constants (e.g., `CreateCode`, `UpdateCode`)
- **messages.go** - payload type definitions (e.g., `CreateCodeData`, `UpdateCodeData`)
- **events.go** - event action constants and payload type aliases
- **entity.go** - entity structure definition

Current entities: `code`, `file`, `project`

### Queries (Return Types)
For each projection in `internal/domain/projections/{entity}-entity/`:
- **view.go** - return type structure for queries
- **query.go** - query-specific logic (if exists)

## Endpoints Structure

### POST /commands
- **Single command**: Accepts one `AnyMessage` with `type: "Command"`
- **Batch commands**: Accepts array of `AnyMessage` objects. On Batch, best effort. See `handlers/http/types` for structure.
- **Request body**: Command with appropriate `aggregateType`, `action`, and `payload`
- **Response**:
  - Single: Returns `AnyMessage` with `type: "DomainEvent"`
  - Batch: Returns `batchResponse` object (see `internal/handlers/http/handlers.go`)
- **Error responses**: 400 (validation), 500 (server error) with `ErrorResponse` format

**Important Command Rules**:
- `aggregateId`: Empty string for CREATE operations, required for UPDATE/DELETE
- `aggregateType`: Must be entity name (capitalized: "Code", "File", "Project")
- `action`: Must match command constant from entity's `commands.go`
- `payload`: Structure from entity's `messages.go`

### GET /queries/{entity}/
Query endpoints defined in `internal/handlers/routes.go`:

### GET /ws/
WebSocket endpoint for bidirectional communication (commands in (same commands as accepted by /commands endpoint), events out (all DomainEvents)). 

## Schema Generation Instructions

1. **Read core CQRS types** from `internal/cqrs/message.go`
2. **For each entity** in `internal/domain/entities/`:
   - Read `commands.go` for command actions
   - Read `messages.go` for payload structures with validation tags
   - Read `entity.go` for complete entity structure
   - Create command examples using realistic qualitative coding data
3. **For query responses**:
   - Read projection `view.go` files for return types
   - Read `internal/handlers/routes.go` for exact endpoint paths
4. **Include all validation constraints** from struct tags (`validate:"required"`, `maxLength`, patterns, etc.)
5. **Create realistic examples** in qualitative coding context:
   - Codes: `topic:climate`, `sentiment:positive`, `method:interview`
   - Files: Interview transcripts, field notes
   - Projects: Research studies, policy analysis
6. **Document error responses** (400, 404, 500) with `ErrorResponse` schema
7. **Include batch command format** and `batchResponse` structure
8. **Describe shapes** for each command, query etc, describe what parameters, return values etc are. Not just the shape, but an explanation of what it represents
 
## Validation Patterns to Include

- **Code slugs**: Pattern `^[a-z0-9-]+:[a-z0-9-]+$` (category:value format, lowercase)
- **Required fields**: Check `validate:"required"` tags in `messages.go` files
- **String lengths**: Check `maxLength` constraints (e.g., title max 200, summary max 1500)
- **Normalization**: Fields with `normalize:"trim,lowercase"` should document expected format
- etc

## Quality Checks

Before writing output:
- [ ] All entities (Code, File, Project, etc) included with their commands
- [ ] All query endpoints from `routes.go` documented
- [ ] Command examples show both single and batch format
- [ ] Payload schemas match exact struct definitions from `messages.go`
- [ ] Chunk query endpoint properly documented with `ChunkResult` response, get chunk size from 
- [ ] Error response format included
- [ ] Batch response format included
- [ ] All validation constraints from struct tags present
- [ ] Examples use qualitative coding domain language

## Colors, where colors

color:
    type: string
    pattern: '^[a-z]+(-[a-z]+)*-(50|100|200|300|400|500|600|700|800|900|950)$'
    example: green-500
    description: Tailwind color class (e.g., emerald-600, light-blue-400)

## Final Step

Write complete specification to `open-api-spec.generated.yml` in repository root using Write tool (override existing).