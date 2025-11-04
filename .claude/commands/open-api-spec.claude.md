# OpenAPI Spec Generation

Generate a complete OpenAPI 3.0 spec for this CQRS-based qualitative data analysis API.

## Output

- **File**: `open-api-spec.generated.yml` (repository root)
- **Name**: hermes-relay-api
- **Version**: 1.0.0
- **Server**: http://localhost:8080
- **Always override**: Use Write tool to replace existing file

## Where to Find Things

### Core Message Structure
**`internal/cqrs/message.go`**
- `Message[T]` - generic wrapper for commands/events
- `AnyMessage` - type-erased version for HTTP
- Fields: `action`, `type`, `aggregateId`, `aggregateType`, `payload`, `timestamp`, `causationId`

### Entities (Commands & Payloads)
**`internal/domain/entities/{entity}/`**

Each entity folder contains:
- **`commands.go`** - command action constants (e.g., `CreateCode`, `UpdateCode`, `DeleteCode`)
- **`messages.go`** - payload structs with validation/normalization tags
- **`events.go`** - event action constants (e.g., `CreatedCode`, `UpdatedCode`, `DeletedCode`)
- **`entity.go`** - entity structure with all fields

**Discovery**: Use `ls internal/domain/entities/` to find all entities, then read each folder.

**Validation Tags**: Look for:
- `validate:"required"` - required fields
- `validate:"min=X,max=Y"` - length/value constraints
- `validate:"gtfield=OtherField"` - field comparison (e.g., end_index > start_index)
- `validate:"pattern"` - custom validation patterns
- `normalize:"trim,lowercase"` - normalization rules

### Query Endpoints & Responses
**`internal/handlers/routes.go`** - ALL endpoint paths and route definitions

**`internal/domain/projections/{entity}-entity/view.go`** - response structs (type aliases to entity)
**`internal/domain/projections/{entity}-entity/query.go`** - custom query types (if exists)

**`internal/projection/queries.go`** - generic query patterns:
- `PaginationResult[T]` - wrapper for paginated responses
- `PaginationQuery` - pagination parameters
- `GetByIDQuery` - ID-based queries
- `EmptyQuery` - list all queries
- Query functions: `Paginate`, `ByID`, `ByAll`, `ThenMap`

### HTTP Layer
**`internal/handlers/http/`**
- **`handlers.go`** - command/query processing, `BatchResponse` format
- **`processors.go`** - batch processing logic
- **`responses.go`** - `batchStatus()` function defines status codes (200/202/207/400)
- **`types.go`** - HTTP-specific types
- **`socket.go`** - WebSocket handler details

### Error Responses
**`internal/lib/utils/errors.go`**
- `ErrorResponse` - standard errors with optional `fields` map
- `NotFoundError` - resource not found (404)
- `ValidationError` - validation errors with field details
- `ConflictError` - conflict errors (409)

### WebSocket
**`internal/handlers/http/socket.go`** - WebSocket implementation details

## API Structure & Behavior

### POST /commands
- **Single**: One `AnyMessage` with `type: "Command"`
- **Batch**: Array of `AnyMessage` objects
- **Returns**:
  - Single: DomainEvent (200) or ErrorResponse (400/500)
  - Batch: `BatchResponse` with status from `batchStatus()` function in `responses.go`
- **Rules**:
  - CREATE: `aggregateId` = `""` (empty string)
  - UPDATE/DELETE: `aggregateId` = entity UUID
  - `aggregateType` must be capitalized entity name

**Batch Processing** (read from `processors.go` and `responses.go`):
- Check `batchStatus()` for exact status code logic
- Typically: all failed → 400, partial success → 207, all success → 200/202
- Each command processed independently (best-effort)
- No cross-command transactions

**Commands without payload structs**: Some commands (like `ClearCoding`, `DeleteCode`) don't have payload structs in `messages.go` because they only need the `aggregateId` from the message. These should be documented with empty or minimal payload requirements.

### GET /queries/{entity}/*
Read all endpoint paths from `routes.go`. Common patterns:
- List endpoints may return `PaginationResult[T]` if they use `Paginate` function
- Single item endpoints return the entity directly
- Custom query endpoints in `{entity}-entity/query.go` may have special return types

**Check `routes.go` for**:
- Which endpoints exist
- Which endpoints use `Paginate` (will return `PaginationResult[T]`)
- Query parameter bindings

**Pagination** (from `projection/queries.go`):
- `PaginationQuery` struct shows accepted parameters
- `PaginationResult[T]` struct shows response format
- Default values should be inferred from validation tags

**Chunks**: If `ChunkResult` type exists in projections:
- Check the struct for exact fields
- Note: chunks use 1-based indexing (first chunk is 1)

### GET /ws/
Read `socket.go` for WebSocket implementation details:
- Message format (JSON)
- Send/receive patterns
- Reconnection/heartbeat behavior
- Error handling

## Command Discovery

**To find all commands**:
1. `ls internal/domain/entities/` to discover entities
2. For each entity, read `commands.go` for command constants
3. For each entity, read `events.go` for event constants
4. Cross-reference with `messages.go` for payload structs
5. For commands without payload structs:
   - Check if command uses only `aggregateId` from the Message (e.g., Delete, Clear operations)
   - These are valid and should be documented as "(no payload required)"
   - Example: "DeleteCode - Delete a code (no payload required)"
   - Example: "ClearCoding - Remove all coding from a file (no payload required)"
6. Mark as "⚠️ Not yet implemented" if:
   - Command constant exists but implementation unclear
   - No payload struct AND no evidence of usage in handlers/tests

Document all found commands in a "Supported Commands by Aggregate" section. For commands without payloads, include what the command does and explicitly note "(no payload required)".

## Validation Patterns

**Extract from struct tags** in `messages.go` and `entity.go` files:
- `validate:"required"` → required field
- `validate:"min=X,max=Y"` → length/value bounds
- `validate:"gtfield=OtherField"` → comparative validation (e.g., end_index > start_index)
- `validate:"gte=0"` → minimum value
- `validate:"code_slug"` or custom validators → check for patterns in comments or validator code
- `normalize:"trim,lowercase"` → automatic normalization

**Pattern Discovery**:
- **Code slugs**: Check code entity `messages.go` for validation pattern
- **Colors**: Check code entity for Tailwind color format and examples
  - Should support multi-word colors (e.g., `light-blue-400`)
  - Should document shade range (50-950)
- **Field lengths**: `maxLength` in OpenAPI = `max` value from validation tags

## Example Domains

Use diverse qualitative research contexts:
- **Healthcare**: patient interviews, clinical observations
- **UX Research**: usability testing, user feedback
- **Policy**: parliamentary debates, impact assessments
- **Education**: classroom observations, student interviews
- **Market Research**: focus groups, consumer interviews
- **Organizational**: employee interviews, workplace studies
- **Social Science**: ethnographic fieldnotes, oral histories

Example code slugs: `topic:patient-experience`, `emotion:anxiety`, `usability:friction-point`, `theme:energy-transition`

## Generation Steps

1. **Core CQRS Types**: Read `internal/cqrs/message.go` for `Message[T]` and `AnyMessage` structure
2. **Discover Entities**: `ls internal/domain/entities/` to find all entity folders
3. **For Each Entity**:
   - Read `commands.go` for command action constants
   - Read `events.go` for event action constants
   - Read `messages.go` for payload structs (check validation/normalization tags)
   - Read `entity.go` for complete entity structure
4. **Endpoints**: Read `routes.go` to discover ALL endpoint paths and their handlers
5. **Projections**: For each entity, check if `internal/domain/projections/{entity}-entity/` exists:
   - Read `view.go` for response types
   - Read `query.go` (if exists) for custom query types
6. **Pagination**: Read `internal/projection/queries.go` for `PaginationResult` and related types
7. **HTTP Layer**:
   - Read `responses.go` for `batchStatus()` to understand status codes
   - Read `processors.go` for batch processing behavior
   - Read `handlers.go` for `BatchResponse` format
   - Read `socket.go` for WebSocket specifics
8. **Error Types**: Read `internal/lib/utils/errors.go` for error response formats
9. **Generate Spec**: Use discovered information to create complete, accurate spec

## Output Requirements

- **Format**: Valid OpenAPI 3.0.3 YAML
- **Accuracy**: All schemas match exact struct definitions from code
- **Validation**: All validation constraints from struct tags included
- **Examples**: Realistic, diverse examples across multiple qualitative research domains
- **Completeness**:
  - All endpoints from `routes.go`
  - All commands/events discovered
  - Error responses (400, 404, 500) with correct schema references
  - Both single and batch command formats
  - WebSocket protocol details from `socket.go`
  - Pagination format from `queries.go`
  - Batch status codes from `responses.go`
- **Documentation**: Clear descriptions explaining what each endpoint/command does
- **Unimplemented Commands**: Mark with ⚠️ when command constant exists but no implementation found