# OpenAPI Spec Generation

Generate a complete OpenAPI 3.0 spec for this CQRS-based qualitative data analysis API.

## Workflow

### Phase 1: DISCOVERY (COMPREHENSIVE SCAN)

**CRITICAL**: ALWAYS perform a complete, thorough scan. Do NOT rely on previous knowledge.

**1. Scan ENTIRE domain folder:**
   - `ls internal/domain/entities/` to discover ALL entities
   - For EACH entity folder, read ALL files:
     - `commands.go` - ALL command constants
     - `events.go` - ALL event constants
     - `messages.go` - ALL payload structs with validation tags
     - `entity.go` - complete entity structure
   - Check for `validators.go` and `handlers.go` for custom validation

**2. Scan routes.go THOROUGHLY:**
   - Read `internal/handlers/routes.go` completely
   - Extract EVERY endpoint path and method
   - Note which query functions are used (Paginate, ByID, ByAll, custom)
   - Identify middleware and special handlers

**3. Scan projections for query details:**
   - For each entity, check `internal/domain/projections/{entity}-entity/`:
     - `view.go` - response types
     - `query.go` - custom query types (if exists)
     - `filter.go` - filter query parameters (CRITICAL - do not skip)
   - Read `internal/cqrs/projection/queries.go` for generic patterns

**4. Extract validation patterns:**
   - From struct tags: `validate:"..."`, `normalize:"..."`
   - From `internal/lib/utils/errors.go`: error message formats

**5. Build complete API model:**
   - All entities with their commands/events/payloads
   - All endpoints with query parameters
   - All validation rules
   - All error response formats

### Phase 2: GENERATE & COMPARE
1. Generate complete spec from discovered model
2. If `open-api-spec.generated.yml` exists, compare with existing spec
3. Categorize changes: ADDED, REMOVED, MODIFIED, UNCHANGED
4. Validate removals (verify they're actually gone from code)
5. Show diff to user in a clear, readable format

### Phase 3: SHOW DIFF & WAIT
- Show the complete diff
- STOP and wait for explicit user confirmation
- Do NOT apply changes unless user explicitly says to apply in a follow-up message
- User must say something like "apply", "yes, apply that", "make those changes", etc.

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

**Struct Tags**: Look for:
- `validate:"required"` - required fields
- `validate:"min=X,max=Y"` - length/value constraints
- `validate:"gtfield=OtherField"` - field comparison (e.g., end_index > start_index)
- `validate:"pattern"` - custom validation patterns
- `normalize:"trim,lowercase"` - normalization rules
- **`query:"paramName"`** - query parameter names (CRITICAL for documenting endpoint query params)

**Custom Validation**: Look for:
- **`validators.go`** - called validation functions
- **`handlers.go`** - Validate calls in route for custom validation (eg duplicate checks, existence checks etc)

### Query Endpoints & Responses
**`internal/handlers/routes.go`** - ALL endpoint paths and route definitions

**`internal/domain/projections/{entity}-entity/view.go`** - response structs (type aliases to entity)
**`internal/domain/projections/{entity}-entity/query.go`** - custom query types (if exists)
**`internal/domain/projections/{entity}-entity/filter.go`** - filter types with query parameters (IMPORTANT: check for these!)

**`internal/projection/queries.go`** - generic query patterns:
- `PaginationResult[T]` - wrapper for paginated responses
- `PaginationQuery` - pagination parameters
- `GetByIDQuery` - ID-based queries
- `EmptyQuery` - list all queries
- Query functions: `Paginate`, `ByID`, `ByAll`, `ThenMap`

**Query Parameters & Filters** (CRITICAL - DO NOT SKIP):
For EACH endpoint in `routes.go`, check for:
1. **`internal/domain/projections/{entity}-entity/filter.go`** - entity-specific filters
2. **`internal/domain/projections/{entity}-entity/query.go`** - custom query structs
3. Look for struct fields with `query:"paramName"` tags
4. Example from chunks: `SearchText string \`query:"searchText"\``
5. All `query:` tagged fields MUST be documented as query parameters in the OpenAPI spec

### HTTP Layer
**`internal/handlers/http/`**
- **`handlers.go`** - command/query processing, `BatchResponse` format
- **`processors.go`** - batch processing logic
- **`responses.go`** - `batchStatus()` function defines status codes (200/202/207/400)
- **`types.go`** - HTTP-specific types
- **`socket.go`** - WebSocket handler details

### Error Responses
**`internal/lib/utils/errors.go`**
- `ValidationError` - validation errors with descriptive messages and field details
  - Format: `{"message": "validation failed: {field} {description}", "fields": {"FieldName": "tag"}}`
  - Message includes human-readable descriptions (e.g., "validation failed: Name is required")
  - Field descriptions defined in `formatFieldError()` function
  - Multiple field errors are comma-separated in message
- `NotFoundError` - resource not found (404)
- `ConflictError` - conflict errors (409)
- `InternalError` - internal server errors (500)

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

**Query Parameters - CRITICAL**:
For EVERY endpoint, check for query parameter structs:
1. Look for `filter.go` in the entity projection folder
2. Look for `query.go` for embedded query/filter structs
3. Extract ALL fields with `query:"paramName"` tags
4. Document each as a query parameter with proper type and description
5. Example: `SearchText string \`query:"searchText"\`` → query param "searchText" of type string

**Pagination** (from `projection/queries.go`):
- `PaginationQuery` struct shows accepted parameters
- `PaginationResult[T]` struct shows response format
- Default values should be inferred from validation tags

**Chunks**: If `ChunkResult` type exists in projections:
- Check the struct for exact fields
- Chunk IDs are opaque identifiers (if no ID provided in query, first chunk is returned)
- IMPORTANT: Check `internal/domain/projections/file-entity/chunk/filter.go` for filter query parameters
- Common chunk filters: searchText, minCoverage, maxCoverage, codeSlugs

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
- **Code slugs**: MUST use this exact pattern in all OpenAPI schema locations: `^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$`
  - Only lowercase letters (a-z)
  - Hyphens (-) for word separation (where spaces would be)
  - Exactly one colon (:) separating category from label
  - NO numbers allowed
  - Examples: `topic:patient-experience`, `emotion:anxiety`, `theme:climate-change`
  - Source: `internal/lib/utils/validate.go` line 22
- **Colors**: Check code entity for Tailwind color format and examples
  - Should support multi-word colors (e.g., `light-blue-400`)
  - Should document shade range (50-950)
- **Field lengths**: `maxLength` in OpenAPI = `max` value from validation tags

### Validation Error Message Formats

**Read `internal/lib/utils/errors.go`** for the `formatFieldError()` function that defines error message templates.

Common validation error messages (based on validation tags):
- `required` → "validation failed: {Field} is required"
- `min=X` → "validation failed: {Field} must be at least X characters"
- `max=X` → "validation failed: {Field} must be at most X characters"
- `code_slug` → "validation failed: {Field} must match code slug format (lowercase with colon and optional dashes)"
  - **IMPORTANT**: Pattern is `^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$` (NO numbers, only lowercase letters and hyphens)
- `lte` → "validation failed: {Field} must not be in the future"
- `gtfield=X` → "validation failed: {Field} must be greater than X"
- `gte=X` → "validation failed: {Field} must be at least X"

**Multiple validation errors** are combined: "validation failed: Color is required, Reasoning is required"

**Example error responses in OpenAPI**:
```yaml
{
  "message": "validation failed: Name is required",
  "fields": {
    "Name": "required"
  }
}
```

```yaml
{
  "message": "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
  "fields": {
    "Slug": "code_slug"
  }
}
```

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

## Discovery Steps (Phase 1)

Always perform complete discovery by reading all code:

1. **Core CQRS Types**: Read `internal/cqrs/commands/message.go` for `Message[T]` and `AnyMessage` structure
2. **Discover Entities**: `ls internal/domain/entities/` to find all entity folders
3. **For Each Entity**:
   - Read `commands.go` for command action constants
   - Read `events.go` for event action constants
   - Read `messages.go` for payload structs (check validation/normalization tags)
   - Read `entity.go` for complete entity structure
   - Read `handlers/handlers.go` for handler structure (check for transforms in ToCreateEntityEvent)
4. **Endpoints**: Read `internal/handlers/routes.go` to discover ALL endpoint paths and their handlers
5. **Projections**: For each entity, check if `internal/domain/projections/{entity}-entity/` exists:
   - Read `view.go` for response types
   - Read `query.go` (if exists) for custom query types
   - **CRITICAL**: Read `filter.go` (if exists) for filter structs with query parameters
   - Extract ALL fields with `query:"paramName"` tags and document as OpenAPI query parameters
6. **Pagination**: Read `internal/cqrs/projection/queries.go` for `PaginationResult` and related types
7. **HTTP Layer**:
   - Read `internal/handlers/http/responses.go` for `batchStatus()` to understand status codes
   - Read `internal/handlers/http/processors.go` for batch processing behavior
   - Read `internal/handlers/http/handlers.go` for `BatchResponse` format
   - Read `internal/handlers/http/socket.go` for WebSocket specifics
8. **Error Types**: Read `internal/lib/utils/errors.go` for:
   - Error response struct definitions
   - `formatFieldError()` function for validation error message templates
   - Error type distinctions (ValidationError, NotFoundError, ConflictError, InternalError)
9. **Validation**: Read `internal/lib/utils/validate.go` for custom validators (e.g., code_slug pattern)

## Comparison Steps (Phase 2)

Always:

1. **Parse Existing Spec**: Read and parse the existing YAML file
2. **Compare Endpoints**:
   - Check each path in existing spec against discovered routes
   - Categorize: ADDED (in code, not in spec), REMOVED (in spec, not in code), MODIFIED (params/schema changed)
3. **Compare Schemas**:
   - Check each schema in existing spec against discovered structs
   - Check fields, validation rules, types
   - Categorize: ADDED, REMOVED, MODIFIED
4. **Compare Parameters**:
   - Check query parameters against filter.go structs
   - Check path parameters against route definitions
5. **Validate Removals**:
   - For each REMOVED item, verify it's actually gone from code
   - Flag if removal seems suspicious (e.g., endpoint exists in routes but flagged as removed)
6. **Build Diff Summary**:
   ```
   ➕ ADDITIONS (X):
     - Path: GET /new/endpoint
     - Schema: NewType
     - Parameter: existingEndpoint?newParam

   ➖ REMOVALS (Y):
     - Path: GET /old/endpoint ✓ VERIFIED (not in routes.go)
     - Schema: OldType ✓ VERIFIED (struct deleted)

   🔄 MODIFICATIONS (Z):
     - Schema File: added fields Type, Original, Locked
     - Path POST /commands: CreateFile payload changed

   ✅ UNCHANGED: N items
   ```

## Generation Steps (Phase 3)

- Show diff summary

## Safety Checks

Before marking anything as REMOVED, verify:
1. **Endpoints**: Check route actually deleted from `routes.go`
2. **Schemas**: Check struct actually deleted from entity files
3. **Parameters**: Check query tag actually removed from filter structs
4. **Commands**: Check command constant actually removed from `commands.go`

If verification fails, don't mark as removed - mark as WARNING instead.

## Diff Output Format

Show clear, actionable diff:

```
OpenAPI Spec Generation - Comparison Summary

DISCOVERY COMPLETE:
  - 3 entities discovered
  - 8 commands found
  - 12 endpoints discovered
  - 45 query parameters found

COMPARING WITH EXISTING SPEC (open-api-spec.generated.yml):

➕ ADDITIONS (5):
  Endpoints:
    + GET /queries/projects/{projectId}/files/{id}/chunks?searchText=string
    + GET /queries/projects/{projectId}/files/{id}/chunks?minCoverage=number
    + GET /queries/projects/{projectId}/files/{id}/chunks?maxCoverage=number
    + GET /queries/projects/{projectId}/files/{id}/chunks?codeSlugs=array

  Schemas:
    + File.Type: FileType enum (codebook|source|memo|context)
    + File.Original: string
    + File.Locked: boolean

➖ REMOVALS (2):
  Endpoints:
    - GET /queries/files/ ✓ VERIFIED (route removed from routes.go)

  Schemas:
    - OldFileType ✓ VERIFIED (type no longer exists)

🔄 MODIFICATIONS (3):
  Schemas:
    ~ File: added 3 fields (Type, Original, Locked)
    ~ CreateFilePayload: Type field now optional (omitempty validation)

  Endpoints:
    ~ POST /commands: CreateFile now sets Type default to 'source'

✅ UNCHANGED: 47 items

Total changes: 10 (5 additions, 2 removals, 3 modifications)
```

## Output Requirements

- **Format**: Valid OpenAPI 3.0.3 YAML
- **Accuracy**: All schemas match exact struct definitions from code
- **Validation**: All validation constraints from struct tags included
  - **CRITICAL**: Code slug pattern MUST be `^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$` in ALL locations (see Pattern Discovery section)
- **Examples**: Realistic, diverse examples across multiple qualitative research domains
- **Completeness**:
  - All endpoints from `routes.go`
  - All commands/events discovered
  - All query parameters from filter.go files
  - Error responses (400, 404, 409, 500) with correct schema references
  - Both single and batch command formats
  - WebSocket protocol details from `socket.go`
  - Pagination format from `queries.go`
  - Batch status codes from `responses.go`
- **Error Examples**: Include realistic validation error examples with descriptive messages:
  - Single field errors: `{"message": "validation failed: Name is required", "fields": {"Name": "required"}}`
  - Multiple field errors: `{"message": "validation failed: Color is required, Reasoning is required", "fields": {"Color": "required", "Reasoning": "required"}}`
  - Pattern validation errors: `{"message": "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)", "fields": {"Slug": "code_slug"}}`
- **Documentation**: Clear descriptions explaining what each endpoint/command does

**IMPORTANT**: This command ONLY shows diffs. It NEVER applies changes automatically.
After showing the diff, ALWAYS wait for explicit user confirmation before applying any changes.
Do NOT assume consent from related discussion - only apply when user explicitly says to apply.

## Output

- **File**: `open-api-spec.generated.yml` (repository root)
- **Name**: hermes-relay-api
- **Version**: 1.0.0
- **Server**: http://localhost:8080