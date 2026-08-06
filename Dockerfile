FROM golang:1.25-alpine AS build

WORKDIR /src

# Their own layer, so editing a source file does not refetch the module cache.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO off is what makes the binary static, which is what allows a scratch final
# stage. -s -w drop the symbol table and DWARF, -trimpath keeps build paths out
# of the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /nabu-storage ./cmd

# Prepared here because the final stage has no shell to mkdir with. An empty
# named volume mounted at /data inherits this ownership, which is what lets the
# unprivileged user below write to it.
RUN mkdir -p /data && chown 65532:65532 /data

FROM scratch

COPY --from=build --chown=65532:65532 /data /data
COPY --from=build /nabu-storage /nabu-storage

# Numeric, because scratch carries no /etc/passwd for a name to resolve against.
USER 65532:65532

EXPOSE 8080

# PERSISTENCE_DIR is deliberately not set. It is where every project directory
# is created, and a wrong guess is a container that serves an empty corpus and
# discards writes on restart, so the server refuses to start until it is given
# an absolute path it can write to. /data is the path this image prepares.
ENTRYPOINT ["/nabu-storage"]

# The binary checks itself, because a scratch image holds one file and no shell,
# no curl and no wget — any other command here would be one the image cannot
# execute, leaving a container permanently unhealthy rather than one that is not
# yet listening.
HEALTHCHECK --interval=5s --timeout=3s --start-period=2s --retries=5 \
    CMD ["/nabu-storage", "healthcheck", "--addr", "127.0.0.1:8080"]
