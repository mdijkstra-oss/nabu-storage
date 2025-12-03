# OpenAPI Spec Generation

Generate a complete OpenAPI 3.0 spec for this CQRS-based qualitative data analysis API.

**Key Principles:**
1. **Adaptive Discovery**: Check what exists, don't assume rigid structure
2. **Consistent Formatting**: Match existing spec formatting exactly (see YAML Formatting Rules)
3. **Complete x-command-reference**: Critical for MCP generation
4. **Wait for Approval**: Show diff, then STOP - don't apply without explicit confirmation

## Workflow

### Phase 1: DISCOVERY (ADAPTIVE SCAN)

**CRITICAL**: ALWAYS perform a complete, thorough scan. Do NOT rely on previous knowledge.

**1. Discover entities (try these in order):**
   - Primary: `ls internal/domain/entities/` to discover entity folders
   - For EACH entity folder, check which files exist (don't assume all exist):
     - `commands.go` - command action constants (if exists)
     - `events.go` - event action constants (if exists)
     - `messages.go` OR `payloads.go` - payload structs with validation tags
     - `entity.go` - entity structure
     - `validators.go` - custom validators (if exists)
     - `handlers/handlers.go` - handler logic and transforms (if exists)
   - If files have different names, adapt - use grep to find command/event constants

**2. Discover routes (try these in order):**
   - Primary: Read `internal/handlers/routes.go` for ALL endpoints
   - Fallback: Check `internal/handlers/http/routes.go`
   - Extract EVERY endpoint path and method
   - Note which query functions are used (Paginate, ByID, ByAll, custom)
   - Identify middleware and special handlers

**3. Discover projections and queries:**
   - For each entity, check if `internal/domain/projections/{entity}-entity/` exists
   - If exists, check for these files (all optional):
     - `view.go` - response types
     - `query.go` - custom query types
     - `filter.go` - filter query parameters (CRITICAL for query params)
   - Read `internal/cqrs/projection/queries.go` for generic patterns (if exists)
   - Look for `query:"paramName"` tags in ANY structs - these become query parameters

**4. Extract validation patterns:**
   - From struct tags: `validate:"..."`, `normalize:"..."`, `query:"..."`
   - From validation files (check these): `internal/lib/utils/errors.go`, `internal/lib/utils/validate.go`
   - Look for custom validator functions

**5. Build complete API model:**
   - All entities with their commands/events/payloads
   - All endpoints with query parameters
   - All validation rules
   - All error response formats

### Phase 2: ANALYZE CODEBASE CHANGES

**Before comparing specs, understand what changed in the code:**

1. **Find last spec update**:
   ```bash
   git log --all -1 --format="%H %s" -- open-api-spec.generated.yml
   ```
   This gives you the commit hash and message of last spec update

2. **Diff codebase since that commit**:
   ```bash
   git diff <commit-hash> HEAD -- internal/domain/entities/ internal/handlers/routes.go internal/domain/projections/
   ```
   Focus on files that affect the API:
   - `internal/domain/entities/*/` - commands, events, payloads
   - `internal/handlers/routes.go` - endpoints
   - `internal/domain/projections/*/filter.go` - query params
   - `internal/lib/utils/validate.go` - validation patterns

3. **Analyze the diff to identify**:
   - New entity files (new commands/events)
   - Deleted entity files (removed commands/events)
   - Modified structs (changed validation, new fields)
   - New routes or removed routes
   - New query parameters in filters

4. **Use this to guide comparison**:
   - If diff shows new `CreateFoo` command, expect it in new spec
   - If diff shows deleted route, expect it removed from new spec
   - If diff shows new validation tag, expect schema change

### Phase 3: GENERATE & COMPARE

1. Generate complete spec from discovered model
2. If `open-api-spec.generated.yml` exists, compare with existing spec
3. Use codebase diff from Phase 2 to validate changes
4. Categorize changes: ADDED, REMOVED, MODIFIED, UNCHANGED
5. Cross-reference with code changes:
   - "Added CreateFoo command ✓ (new in commit abc123)"
   - "Removed OldEndpoint ✓ (deleted in commit def456)"
   - "Modified File schema ✓ (validation changed in commit ghi789)"
6. Show diff to user in a clear, readable format

### Phase 4: SHOW DIFF & WAIT
- Show the complete diff
- STOP and wait for explicit user confirmation
- Do NOT apply changes unless user explicitly says to apply in a follow-up message
- User must say something like "apply", "yes, apply that", "make those changes", etc.

## YAML Formatting Rules

**CRITICAL**: Match the exact formatting style of the existing spec. Consistency is key for MCP generation.

### Indentation & Spacing
- **Indentation**: 2 spaces (never tabs)
- **Section spacing**: Single empty line between major sections (paths, components, each schema)
- **Property spacing**: No empty lines between properties within a schema
- **Example spacing**: No empty line before `example:` or `examples:`

### Descriptions
- **Multi-line descriptions**: Use pipe `|` with proper indentation
  ```yaml
  description: |
    First line of description.

    Second paragraph after blank line.
  ```
- **Single-line descriptions**: Inline string (no pipe)
  ```yaml
  description: Brief description here
  ```
- **When to use multi-line**: Use pipe if description has multiple paragraphs or needs line breaks

### Schema Ordering
**Top-level schemas in `components/schemas`**: Alphabetical order
- AnyMessage
- BatchResponse
- Chunk
- ChunkResult
- Code
- CodePaginationResult
- (etc.)

**Within each schema**, order properties semantically:
1. `type` (always first)
2. `required` (if present)
3. `properties` (main content)
4. `enum` (if present)
5. `description`, `example`, etc. (metadata last)

**Within `properties`**, order fields logically:
1. ID fields first (`id`, `project_id`, etc.)
2. Core fields (`name`, `slug`, etc.)
3. Descriptive fields (`description`, `reasoning`, etc.)
4. Collection fields (`code_ids`, `file_ids`, arrays, etc.)
5. Metadata fields (`timestamp`, `version`, etc.)

### Examples
- **Named examples**: Use when multiple examples needed
  ```yaml
  examples:
    exampleName:
      summary: Brief description
      value:
        field: value
  ```
- **Inline examples**: Use for simple single examples
  ```yaml
  example: simple value
  ```
- **Example values**: Use realistic qualitative research data (healthcare, UX, policy, etc.)

### Enums & Arrays
- **Inline arrays**: `[value1, value2, value3]` for short lists
- **Multi-line arrays**: Only if list is very long or values are complex
  ```yaml
  enum:
    - longValue1
    - longValue2
  ```

### References
- Always use `$ref: '#/components/schemas/TypeName'`
- Keep schema names in PascalCase

### Query Parameters
Each parameter as separate object:
```yaml
parameters:
  - name: paramName
    in: query
    description: What this param does
    schema:
      type: string
      default: value
```

### Response Examples
Use named examples with clear summaries:
```yaml
responses:
  '200':
    description: Success response
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Type'
        examples:
          exampleName:
            summary: Clear description of this example
            value:
              field: value
```

### x-command-reference Section
**CRITICAL**: This section is used for MCP generation. Format exactly as:
```yaml
x-command-reference:
  description: |
    Complete reference of all supported commands by aggregate type.

  EntityName:
    CommandName:
      description: What this command does
      payload: PayloadTypeName
      aggregateId: empty string for create | entity UUID for update
      event: EventName
      validation:
        - First validation rule
        - Second validation rule
      note: Additional notes (optional)
      example: |
        {
          "type": "Command",
          "action": "CommandName",
          ...
        }
```

### Property Formatting
- **String types**: Include `minLength`, `maxLength`, `pattern` from validation tags
- **Number types**: Include `minimum`, `maximum`, `format` (int32, double, etc.)
- **Arrays**: Specify `minItems`, `maxItems` if validated
- **Format hints**: Use `format: uuid`, `format: date-time`, etc. where applicable

### Validation in Schemas
Translate Go struct tags to OpenAPI:
- `validate:"required"` → `required: [field]`
- `validate:"min=X,max=Y"` → `minLength: X, maxLength: Y`
- `validate:"pattern"` → `pattern: 'regex'`
- `validate:"gte=X"` → `minimum: X`
- `validate:"lte=X"` → `maximum: X`

## Where to Find Things (Typical Locations)

These are the typical locations. If files don't exist at these paths, search for them or use alternative discovery methods.

### Core Message Structure
**Check these locations:**
- `internal/cqrs/commands/message.go` OR `internal/cqrs/message.go`
- Look for: `Message[T]`, `AnyMessage`, core message fields

### Entities (Commands & Payloads)
**Primary location**: `internal/domain/entities/{entity}/`

**Typical files in each entity folder (check which exist):**
- `commands.go` - command action constants (e.g., `CreateCode`, `UpdateCode`)
- `events.go` - event action constants (e.g., `CreatedCode`, `UpdatedCode`)
- `messages.go` OR `payloads.go` - payload structs with validation tags
- `entity.go` - entity structure with all fields
- `handlers/handlers.go` - command handlers and transformations (optional)
- `validators.go` - custom validators (optional)

**Discovery strategy:**
1. Use `ls internal/domain/entities/` to find entity folders
2. For each entity, check which files exist (don't assume all are present)
3. If expected files missing, use grep to find constants/structs

**Struct Tags to extract:**
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
**Routes location (check these):**
- `internal/handlers/routes.go` (primary)
- `internal/handlers/http/routes.go` (fallback)

**Projection locations (check if exist):**
- `internal/domain/projections/{entity}-entity/view.go` - response types
- `internal/domain/projections/{entity}-entity/query.go` - custom query types
- `internal/domain/projections/{entity}-entity/filter.go` - **CRITICAL** for query parameters
- `internal/domain/projections/{entity}-entity/chunk/filter.go` - chunk-specific filters

**Generic query patterns (check these):**
- `internal/cqrs/projection/queries.go` OR `internal/projection/queries.go`
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
**Check these locations:**
- `internal/handlers/http/handlers.go` - command/query processing
- `internal/handlers/http/processors.go` - batch processing logic
- `internal/handlers/http/responses.go` - status code logic (`batchStatus()` function)
- `internal/handlers/http/types.go` - HTTP-specific types (e.g., `BatchResponse`)
- `internal/handlers/http/socket.go` - WebSocket implementation

**What to extract:**
- Batch response format
- Status code logic (200/202/207/400)
- WebSocket message format

### Error Responses
**Check these files:**
- `internal/lib/utils/errors.go` (primary)
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
- **Colors**: Based on `radix_color` validation in `internal/lib/utils/validate.go`
  - Radix base colors only (no step numbers): gray, mauve, slate, sage, olive, sand, tomato, red, ruby, crimson, pink, plum, purple, violet, iris, indigo, blue, cyan, teal, jade, green, grass, bronze, gold, brown, orange, amber, yellow, lime, mint, sky
  - Format: just the color name (e.g., `red`, `blue`, `amber`)
  - Use enum in OpenAPI: `enum: [gray, mauve, slate, sage, olive, sand, tomato, red, ruby, crimson, pink, plum, purple, violet, iris, indigo, blue, cyan, teal, jade, green, grass, bronze, gold, brown, orange, amber, yellow, lime, mint, sky]`
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

Always perform complete discovery. Be adaptive - check if files exist before trying to read them.

1. **Core CQRS Types**:
   - Try: `internal/cqrs/commands/message.go`, `internal/cqrs/message.go`
   - Extract: `Message[T]`, `AnyMessage`, core fields

2. **Discover Entities**:
   - Primary: `ls internal/domain/entities/`
   - Fallback: Search for entity definitions with grep

3. **For Each Entity** (be adaptive about which files exist):
   - Check for and read: `commands.go`, `events.go`
   - Check for and read: `messages.go` OR `payloads.go`
   - Check for and read: `entity.go`
   - Optionally read: `handlers/handlers.go`, `validators.go`
   - Extract all struct tags: `validate:`, `normalize:`, `json:`

4. **Discover Endpoints**:
   - Try: `internal/handlers/routes.go`, `internal/handlers/http/routes.go`
   - Extract: ALL paths, methods, query function calls

5. **Projections** (for each entity, check what exists):
   - Look for: `internal/domain/projections/{entity}-entity/`
   - If exists, check for: `view.go`, `query.go`, `filter.go`
   - **CRITICAL**: Extract ALL `query:"paramName"` tags from any structs
   - These become OpenAPI query parameters

6. **Pagination**:
   - Try: `internal/cqrs/projection/queries.go`, `internal/projection/queries.go`
   - Extract: `PaginationResult[T]`, `PaginationQuery` structures

7. **HTTP Layer** (check what exists):
   - `internal/handlers/http/responses.go` - status code logic
   - `internal/handlers/http/processors.go` - batch behavior
   - `internal/handlers/http/handlers.go` - response formats
   - `internal/handlers/http/socket.go` - WebSocket details

8. **Error Handling**:
   - Try: `internal/lib/utils/errors.go`
   - Extract: Error response formats, `formatFieldError()` templates

9. **Validation**:
   - Try: `internal/lib/utils/validate.go`
   - Extract: Custom validators (especially code_slug pattern)

## Comparison Steps (Phase 3)

Always perform comparison with git context:

1. **Parse Existing Spec**: Read and parse the existing YAML file

2. **Get Git Context** (from Phase 2):
   - Last spec commit hash
   - List of files changed since then
   - Specific changes (new commands, modified structs, etc.)

3. **Compare Endpoints**:
   - Check each path in existing spec against discovered routes
   - Categorize: ADDED (in code, not in spec), REMOVED (in spec, not in code), MODIFIED (params/schema changed)
   - Cross-reference with git diff: "Added GET /foo ✓ (route added in commit abc123)"

4. **Compare Schemas**:
   - Check each schema in existing spec against discovered structs
   - Check fields, validation rules, types
   - Categorize: ADDED, REMOVED, MODIFIED
   - Cross-reference: "Added Description field ✓ (struct modified in commit def456)"

5. **Compare Commands in x-command-reference**:
   - Check each command in spec against discovered commands
   - Cross-reference: "Added UpdateProject ✓ (command added in commit ghi789)"

6. **Compare Parameters**:
   - Check query parameters against filter.go structs
   - Check path parameters against route definitions

7. **Validate Removals**:
   - For each REMOVED item, check git diff confirms deletion
   - Flag if git diff shows item still exists
   - Flag if removal seems suspicious

8. **Build Diff Summary with Git Context**:
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

Show clear, actionable diff with git context:

```
OpenAPI Spec Generation - Comparison Summary

DISCOVERY COMPLETE:
  - 3 entities discovered
  - 8 commands found
  - 12 endpoints discovered
  - 45 query parameters found

CODEBASE CHANGES (since last spec update):
  Last spec commit: fac786d docs: update open api spec
  Commits since: 3 commits

  Code changes detected:
    + internal/domain/entities/project/commands.go: UpdateProject, DeleteProject
    + internal/domain/entities/file/commands.go: UpdateFile
    ~ internal/domain/entities/project/messages.go: Description field added
    ~ internal/domain/entities/file/entity.go: Description field added
    ~ internal/cqrs/projection/reducer.go: IfExists wrapper added (no API impact)

COMPARING WITH EXISTING SPEC (open-api-spec.generated.yml):

➕ ADDITIONS (6):
  Commands:
    + UpdateProject ✓ (added in commit 8f8d780)
    + DeleteProject ✓ (added in commit 8f8d780)
    + UpdateFile ✓ (added in commit deec6e8)

  Schemas:
    + UpdateProjectPayload (Name, Description fields)
    + UpdateFilePayload (Name, Description fields)
    + Project.Description field
    + File.Description field

➖ REMOVALS (0):
  (No removals detected)

🔄 MODIFICATIONS (2):
  Schemas:
    ~ Project: added Description field ✓ (commit 8f8d780)
    ~ File: added Description field ✓ (commit deec6e8)

  x-command-reference:
    ~ Project section: added UpdateProject, DeleteProject entries
    ~ File section: added UpdateFile entry

✅ UNCHANGED: 52 items

Total changes: 8 (6 additions, 0 removals, 2 modifications)

All changes align with codebase changes since last spec update.
```

## x-command-reference Section (CRITICAL FOR MCP)

**This section is used to generate MCP (Model Context Protocol) tools. It MUST be complete and correctly formatted.**

### Format Requirements

```yaml
x-command-reference:
  description: |
    Complete reference of all supported commands by aggregate type.

    For aggregateId field:
    - CREATE operations: Use empty string ""
    - UPDATE/DELETE operations: Use entity UUID

  EntityName:
    CommandName:
      description: Clear description of what this command does
      payload: PayloadSchemaName
      aggregateId: empty string for create | entity UUID for update/delete
      event: EventName
      validation:  # Optional - only if there are special validation rules
        - First validation rule
        - Second validation rule
      note: Additional notes  # Optional
      example: |
        {
          "type": "Command",
          "action": "CommandName",
          "aggregateType": "EntityName",
          "aggregateId": "",
          "payload": {
            "field": "value"
          }
        }
```

### Content Requirements

**For EACH discovered command:**
1. Group by entity (Project, Code, File, etc.)
2. Include ALL commands found in `commands.go` files
3. Map command to its payload schema (from components/schemas)
4. Show resulting event name (from `events.go`)
5. List any special validation rules
6. Provide complete example with realistic data

**Commands without payloads:**
- Some commands (DeleteCode, ClearCoding) may not need payload
- Still document them with `payload: none required` or empty payload object
- Example should show minimal structure

**Order:**
- Group commands by entity
- Within each entity, order: Create, Update, Delete, then other commands
- Match capitalization from code (EntityName, CommandName)

## Adaptive Discovery Philosophy

**Be flexible, not rigid:**
- Check if files exist before reading them
- If expected file doesn't exist, search for alternatives
- Don't assume all entities have all file types
- Use grep/search when direct paths don't work
- Adapt to project structure changes

**Report what you find:**
- "Found commands.go in entity X"
- "No messages.go found, checked for payloads.go - found"
- "filter.go exists, extracting query parameters"
- "handlers/handlers.go not found, skipping transforms"

**Still be thorough:**
- Check all typical locations
- Look for patterns across entities
- Extract all validation tags
- Find all query parameters
- Document everything discovered

## Critical Patterns & Validation

### Code Slug Pattern (CRITICAL - Gets Wrong Often)

**The ONLY correct pattern for code slugs is:**
```
^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$
```

**Rules:**
- ONLY lowercase letters `a-z` (NO uppercase, NO numbers)
- Hyphens `-` allowed for word separation (multi-word slugs)
- Exactly ONE colon `:` separating category from label
- Format: `category:label` or `multi-word-category:multi-word-label`

**Valid examples:**
- `emotion:anxiety`
- `topic:patient-experience`
- `usability:friction-point`
- `theme:climate-change`
- `speaker:mark-rutte`

**Invalid (DO NOT ACCEPT):**
- `Emotion:Anxiety` (uppercase)
- `emotion-anxiety` (no colon)
- `emotion:anxiety:detail` (multiple colons)
- `emotion:123` (numbers not allowed)
- `emotion_anxiety` (underscore not allowed)

**Where to use this pattern:**
- `Code.slug` schema property
- `CreateCodePayload.slug` schema property
- `UpdateCodePayload.slug` schema property
- `CodedSection.code_slug` schema property
- `CodingAction.code_slug` schema property
- Path parameter validation in `/queries/projects/{projectId}/codes/slug/{slug}`
- ALL examples using code slugs

**Source**: `internal/lib/utils/validate.go` line 22

### Validation Tag Translations

Always translate Go validation tags to OpenAPI constraints:
- `validate:"required"` → `required: [fieldName]`
- `validate:"min=X,max=Y"` → `minLength: X, maxLength: Y` (for strings)
- `validate:"gte=X,lte=Y"` → `minimum: X, maximum: Y` (for numbers)
- `validate:"code_slug"` → `pattern: '^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$'`
- `validate:"uuid"` → `format: uuid`
- `json:"field_name"` → property name is `field_name`

## Output Requirements

- **Format**: Valid OpenAPI 3.0.3 YAML matching the formatting rules above
- **Accuracy**: All schemas match exact struct definitions from code
- **Validation**: All validation constraints from struct tags included
  - **CRITICAL**: Code slug pattern MUST be `^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$` in ALL locations
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
  - **CRITICAL**: Complete `x-command-reference` section (used for MCP generation)
- **Error Examples**: Include realistic validation error examples with descriptive messages:
  - Single field errors: `{"message": "validation failed: Name is required", "fields": {"Name": "required"}}`
  - Multiple field errors: `{"message": "validation failed: Color is required, Reasoning is required", "fields": {"Color": "required", "Reasoning": "required"}}`
  - Pattern validation errors: `{"message": "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)", "fields": {"Slug": "code_slug"}}`
- **Documentation**: Clear descriptions explaining what each endpoint/command does

**IMPORTANT**: This command ONLY shows diffs. It NEVER applies changes automatically.
After showing the diff, ALWAYS wait for explicit user confirmation before applying any changes.
Do NOT assume consent from related discussion - only apply when user explicitly says to apply.

## Entity Schema Files

**In addition to the OpenAPI spec, update JSON schema files for each entity.**

### Location
Each entity has a `schema.json` file in its folder:
- `internal/domain/entities/code/schema.json`
- `internal/domain/entities/file/schema.json`
- `internal/domain/entities/project/schema.json`

### Discovery
1. `ls internal/domain/entities/` to find all entity folders
2. For each entity, check if `schema.json` exists
3. Read `entity.go` to get current struct definition
4. For Project: use `ProjectArrayView` from `internal/domain/projections/project-entity/view.go` (arrays not maps)

### Schema Generation Rules
1. **$schema**: Always `http://json-schema.org/draft-07/schema#`
2. **$comment**: Include LLM guidance (e.g., "LLM: Keep in sync with entity.go")
3. **For Project specifically**: Comment must note it uses array children and reference `ProjectArrayView`
4. **required**: Extract from `validate:"required"` tags
5. **properties**: Map Go types to JSON Schema types
   - `string` → `type: "string"`
   - `int` → `type: "integer"`
   - `bool` → `type: "boolean"`
   - `time.Time` → `type: "string", format: "date-time"`
   - `[]T` → `type: "array", items: {...}`
   - Embedded structs → flatten properties
6. **validation**: Translate tags
   - `min=X,max=Y` → `minLength/maxLength` or `minimum/maximum`
   - `oneof=a b c` → `enum: ["a", "b", "c"]`
   - Custom validators (e.g., `radix_color`) → appropriate enum
7. **nested types**: Use `$defs` for types like Chunk, CodedSection, Actor
8. **references**: Project schema references code and file schemas via `$ref`

### Workflow Integration
After Phase 3 (Generate & Compare OpenAPI):
1. For each entity folder found in discovery
2. Generate schema.json content from entity.go (or ProjectArrayView for project)
3. Compare with existing schema.json (if exists)
4. Include in diff output under "Entity Schemas" section
5. Create new schema.json files for new entities

### Diff Output for Entity Schemas
```
ENTITY SCHEMAS:

📄 code/schema.json:
  🔄 MODIFIED: Added counter_examples maxItems constraint

📄 file/schema.json:
  ✅ UNCHANGED

📄 project/schema.json:
  ➕ NEW FIELD: phase enum updated

📄 newentity/schema.json:
  ➕ NEW FILE (entity discovered, no schema exists)
```

### Example Schema Structure
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$comment": "LLM: Keep in sync with entity.go Code struct",
  "title": "Code",
  "type": "object",
  "required": ["id", "project_id", "slug", "color", "definition"],
  "properties": {
    "id": { "type": "string" },
    "slug": {
      "type": "string",
      "pattern": "^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$"
    }
  }
}
```

## Output

- **OpenAPI File**: `open-api-spec.generated.yml` (repository root)
- **Entity Schemas**: `internal/domain/entities/{entity}/schema.json` (one per entity)
- **Name**: hermes-relay-api
- **Version**: 1.0.0
- **Server**: http://localhost:8080