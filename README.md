# nabu-storage

A small Go server that stores flat markdown files per project and sends them to a client over WebSocket. A project is a UUID, and its files live flat in one directory — no subdirectories. It is scaffolding for a frontend editor: enough of a backend to make the client work end to end.

> [!WARNING]
> This is unfinished, and there is no authentication. The project UUID is the only credential: anyone holding one can read, write, delete, and rename every file in that project. The WebSocket upgrade accepts any origin and CORS defaults to `*`. Run it on localhost or behind something that authenticates.

## Running

```bash
cp .env.example .env   # fill in PERSISTENCE_DIR
make start
```

| Variable | Default | Meaning |
| --- | --- | --- |
| `PERSISTENCE_DIR` | none — required | Absolute path to a directory that exists and is writable; project directories are created under it |
| `PORT` | `8080` | Listen port |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `CORS_ORIGINS` | `*` | Comma-separated allowed origins |

A leading `~` in `PERSISTENCE_DIR` is expanded; anything else relative is refused. The check runs at startup and the server exits when it fails, naming what it rejected:

```text
{"time":"…","level":"ERROR","msg":"configuration rejected","error":"PERSISTENCE_DIR \"/srv/nabu\": directory does not exist"}
```

### Docker

```bash
docker compose up
```

That publishes `8080` on loopback and keeps the projects in a named volume; `STORAGE_PORT` moves the host side, and the container listens on `8080` either way.

The image leaves `PERSISTENCE_DIR` unset. `compose.yaml` points it at `/data`, and `docker run` has to do the same:

```bash
docker run -p 8080:8080 -e PERSISTENCE_DIR=/data -v nabu-data:/data nabu-storage
```

The image prepares `/data` owned by UID 65532, the unprivileged user the process runs as, which is what makes an empty named volume mounted there writable. A bind mount keeps the host's own ownership instead, so that directory has to be writable by 65532 before the container will accept it.

`SIGINT` or `SIGTERM` stops the listener and exits `0`, giving requests already in flight five seconds to finish. Open websockets do not hold it up.

## API

Every project endpoint takes a `projectId` that must be a UUID. Any spelling `uuid.Parse` accepts is normalised to canonical form, so `{A1B2…}` and `urn:uuid:a1b2…` address one project.

### `POST /commands/{projectId}`

Applies one command.

```json
{ "action": "WriteFile", "path": "notes.md", "content": "# Notes" }
```

The accepted actions are `WriteFile`, `DeleteFile`, and `RenameFile`, the last taking a `newPath` alongside `path`. Anything else is a 400. `WriteFile` creates the project directory when it is not already there.

`WriteFile` replaces the whole file, writing a complete one beside it and renaming it into place. Rename is atomic within a directory, so another process reading the project directory sees either the old file or the new one, never a half-written one, and a write that fails leaves the previous content untouched.

A `path` is a single file name: a leading `.`, a `..` anywhere, and any character outside letters, digits and `-_. (),'` are rejected. That is what keeps a project flat.

Bodies may be sent with `Content-Encoding: gzip`. They are capped at 8 MiB on the wire and 32 MiB decompressed, and either ceiling returns 413.

### `GET /ws/{projectId}`

Upgrades, then immediately sends the whole project as newline-delimited JSON: a count, then one frame per file in alphabetical order.

```json
{"action":"SyncMeta","fileCount":2}
{"action":"WriteFile","path":"notes.md","content":"# Notes"}
{"action":"WriteFile","path":"preferences.md","content":"# Preferences\n\n"}
```

`fileCount` counts the frames that follow, so a file the server cannot read is left out of both.

Those frames are everything the connection carries. Nothing is pushed afterwards, and a command another client posts does not arrive here — a client that wants later writes reconnects. Messages the client sends are read and discarded; ping and pong keep the connection open.

`preferences.md` and `settings.hidden.md` are seeded from templates when a project directory exists without them. Connecting to a UUID that has never been written to reads nothing and creates nothing.

### `GET /queries/projects`

Lists the projects, most recently written first.

```json
{
  "items": [{ "id": "a1b2…", "updatedAt": "2026-08-06T14:57:03.361255521+02:00" }],
  "total": 1,
  "page": 1,
  "page_size": 50
}
```

`page` and `page_size` are optional; anything unparseable or below one falls back to the default, and `page_size` is capped at 200. A page past the end is an empty page rather than a 404.

The answer is read from `PERSISTENCE_DIR` at request time, so a project directory created by hand is listed like any other. A directory whose name is not a UUID is skipped, because neither other endpoint could open it. `updatedAt` is the newest file in the project rather than the directory's own timestamp, which does not move when a file is rewritten in place.

### `GET /health`

Answers `ok` while the process is serving, and is what the container's `HEALTHCHECK` calls — the binary calling itself, since the image is built from `scratch` and holds no curl. It reports on this process alone: the persistence directory was settled at startup, and a server that failed that check is not running to be asked.

## Development

```bash
make setup   # install lefthook, gotestsum, golangci-lint
make dev     # rebuild and restart on change
make test    # or test-race, coverage
```

> [!NOTE]
> `make dev` needs watchexec, which is not a Go package: `brew install watchexec`.

Send a single command against a running server:

```bash
make submit PROJECT_ID=<uuid> JSON='{"action":"WriteFile","path":"a.md","content":"hi"}'
```

## License

AGPL-3.0. Running a modified version as a network service obliges you to offer its source to the people using it.
