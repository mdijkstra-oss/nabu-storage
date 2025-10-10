
# Event-Driven Architecture Patterns

## HTTP Response Handling

### Command vs Event Endpoints

There are two types of HTTP endpoints:

- **`/command`** - Executes a command and returns the business result (200/201 with response body)
- **`/event`** - Accepts an event notification and acknowledges receipt (200 with no body, or error)

### First-Come-First-Serve Pattern

Only the **first subscriber** that returns a non-nil `*Message` will have its result sent back to the HTTP client on the `/command` endpoint. This is by design.

**Convention:**
- **Command handlers** (AddUser, CreateOrder, etc.) should return a response - typically only ONE handler per command
- **Event handlers** (UserAdded, OrderCreated, etc.) should return `nil` - these are side effects that don't respond to the client

```go
// Command handler - returns response
publisher.Subscribe(func(ctx context.Context, msg *Message, pub PublishFunc) (*Message, error) {
if msg.Type != "AddUser" {
return nil, nil
}

    user := createUser(msg.Data)
    
    // Publish event for other handlers
    pub(ctx, &Message{Type: "UserAdded", Data: user})
    
    // Return response to HTTP client
    return &Message{Type: "UserAddedResponse", Data: user}, nil
})

// Side effect handler - no response
publisher.Subscribe(func(ctx context.Context, msg *Message, pub PublishFunc) (*Message, error) {
if msg.Type != "UserAdded" {
return nil, nil
}
sendWelcomeEmail(msg.Data["email"])
return nil, nil  // Don't return to client
})
```

## HTTP Status Codes

### Command Endpoint Status Codes

Messages with types ending in `"Created"` or `"Added"` automatically return `201 Created` status code. All other successful responses return `200 OK`.

```go
// These return 201 Created
UserCreated
OrderCreated  
ProductAdded
TeamAdded

// These return 200 OK
UserUpdated
OrderCancelled
ProductListed
```

### Event Endpoint Status Codes

The `/event` endpoint only returns:
- `200 OK` - Event accepted and processed successfully
- `400 Bad Request` - Malformed request
- `500 Internal Server Error` - Processing failed

## Error Handling

### Custom Error Types

Return specific error types from your handlers to get appropriate HTTP status codes:

```go
// 400 Bad Request
return nil, &ValidationError{Message: "Email is required"}

// 404 Not Found
return nil, &NotFoundError{Resource: "User"}

// 409 Conflict
return nil, &ConflictError{Message: "Email already exists"}

// 500 Internal Server Error
return nil, fmt.Errorf("database connection failed")
```

**Available error types:**
- `ValidationError` → 400 Bad Request
- `NotFoundError` → 404 Not Found
- `ConflictError` → 409 Conflict
- Any other error → 500 Internal Server Error

### WebSocket Error Handling

WebSocket errors are sent as `ErrorResponse` objects:

```json
{
  "error": "Email already exists",
  "type": "conflict",
  "originalMessage": "AddUser"
}
```