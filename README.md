# nabu-storage

A small Go server that stores flat markdown files per project and streams them to a client over WebSocket.

It exists as scaffolding for a frontend editor — enough of a backend to make the client work end to end. It is not a finished product, and the gaps below are load-bearing, not oversights.

> [!WARNING]
> There is no authentication. The project UUID is the only credential: anyone holding one can read, write, delete, and rename every file in that project. The WebSocket upgrade accepts any origin and CORS defaults to `*`. Run it on localhost or behind something that authenticates.

## Running

```bash
cp .env.example .env   # fill in PERSISTENCE_DIR
make start
```

| Variable | Default | Meaning |
| --- | --- | --- |
| `PORT` | `8080` | Listen port |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `PERSISTENCE_DIR` | none — required | Where project directories are created |
| `CORS_ORIGINS` | `*` | Comma-separated allowed origins |

`PERSISTENCE_DIR` must name a directory that already exists and accepts a write, by absolute path — a leading `~` is expanded, anything else relative is refused. The check runs at startup and the server exits when it fails, naming what it rejected:

```text
{"time":"…","level":"ERROR","msg":"configuration rejected","error":"PERSISTENCE_DIR \"/srv/nabu\": directory does not exist"}
```

A server that started against a path that was not there would answer every read with an empty project and drop every write, and would do it quietly.

### Docker

```bash
docker compose up
```

That publishes `8080` on loopback and keeps the projects in a named volume. `STORAGE_PORT` moves the host side; the container listens on `8080` either way.

The image is built from `scratch` and holds the binary, so there is nothing in it to shell into — the `HEALTHCHECK` is the binary asking itself. It prepares `/data` owned by the unprivileged user the process runs as, which is what makes an empty named volume mounted there writable. A bind mount keeps the host's own ownership instead, so that directory has to be writable by UID 65532 before the container will accept it.

`PERSISTENCE_DIR` is not set in the image. `compose.yaml` points it at `/data`, and `docker run` has to do the same:

```bash
docker run -p 8080:8080 -e PERSISTENCE_DIR=/data -v nabu-data:/data nabu-storage
```

## API

The two project endpoints take a `projectId` that must be a UUID. Any spelling `uuid.Parse` accepts is normalised to canonical form, so `{A1B2…}` and `urn:uuid:a1b2…` address one project. Files live flat in `$PERSISTENCE_DIR/{projectId}/` — no subdirectories.

**`POST /commands/{projectId}`** applies one command. Accepts `Content-Encoding: gzip`.

```json
{ "action": "WriteFile", "path": "notes.md", "content": "# Notes" }
```

The accepted actions are `WriteFile`, `DeleteFile`, and `RenameFile`, the last taking a `newPath`. Anything else is a 400.

> [!NOTE]
> Bodies are capped at 8 MiB on the wire and 32 MiB decompressed; either ceiling returns 413.

**`GET /ws/{projectId}`** upgrades and immediately sends the whole project as newline-delimited JSON: a count, then one frame per file in alphabetical order.

```json
{"action":"SyncMeta","fileCount":2}
{"action":"WriteFile","path":"notes.md","content":"# Notes"}
{"action":"WriteFile","path":"preferences.md","content":"# Preferences\n\n"}
```

`fileCount` counts the frames that follow, so a file the server cannot read is left out of both.

`preferences.md` and `settings.hidden.md` are seeded from templates when a project directory exists without them. Connecting to a UUID that has never been written to reads nothing and creates nothing.

**`GET /queries/projects`** lists the projects, most recently written first.

```json
{
  "items": [{ "id": "a1b2…", "updatedAt": "2026-08-06T14:57:03.361255521+02:00" }],
  "total": 1,
  "page": 1,
  "page_size": 50
}
```

`page` and `page_size` are optional; anything unparseable or below one falls back to the default, and `page_size` is capped at 200. A page past the end is an empty page rather than a 404.

The answer is read from `$PERSISTENCE_DIR` at request time, so a directory created by hand is listed like any other. A directory whose name is not a UUID is skipped, because neither other endpoint could open it. `updatedAt` is the newest file in the project rather than the directory's own timestamp, which does not move when a file is rewritten in place.

**`GET /health`** answers `ok` while the process is serving, and is what the container's `HEALTHCHECK` calls — the binary calling itself, since the image holds no curl. It reports on this process alone: the persistence directory was settled at startup, and a server that failed that check is not running to be asked.

## What isn't built

- **Sync is one-directional.** A write over HTTP is not pushed to connected clients. Reconnecting is the only way to pick up another device's changes.
- **`SyncMeta` is send-only.** The server emits it but has no handler for it, so a client that POSTs one gets a 400.
- **Nothing is versioned.** A write overwrites; there is no history and no way back.

## The idea it was heading toward

Because a project is almost entirely markdown and other plain text, each one could be a git repository rather than a plain directory.

The server would own the repo and generate commits itself, grouping incoming commands into units of work rather than committing per keystroke. Users would get history and rollback without ever seeing git: a timeline of their own edits, and a way back to any point on it.

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
