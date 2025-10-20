# OpenAPI Spec
You should generate an open api spec for this repository according to the following definitions

- Name: hermes-relay-api
- servers: http://localhost:8080
- version: 1

These are the endpoints for a CQRS system. Build in Go. When giving samples. Create them where they make sense in the context of Qualitative Coding. 

Write output to `open-api-spec.generated.yml` in root (override existing if needed)

## Authentication
As of now, no authentication

## Endpoints

### /commands/
Look for the `Message` type in this repository, every command follows this format. `Message.Type` always is of the Command type

In the `domain/entities` folder, you can find the entities for which commands can be exposed.
For each command set the appropriate `AggregateType` to the entity (as per struct name, eg Uppercase) and AggregateId should be required when it makes sense in context (eg updates, deletes etc).

Sample for AggregateType `Code`
Sample for action `Command`

Each entity has a `commands.go` file. From there you can find the commands to apply. 
Be sure to also check the corresponding entity definition in the `view.go` folder. 

Supports both single and batch commands

---
For now just assume proper header responses for errors (400, 500 etc) and 200 range for success. (only 200 and 201 atm)
---

### /queries/

For each entity if there is a `query.go` file there is a way to query them. Return value is the `view.go` format.

The format for the endpoint will be lowercased-snakerized entity eg `/queries/file/`

For now only two features on queries:

- fetch all (no pagination)
- fetch by id