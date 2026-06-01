# syntax=docker/dockerfile:1

# -----------------------------------------------------------------------------
# Build stage — compiles all binaries with a pinned Go toolchain.
# -----------------------------------------------------------------------------
FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=0.0.1
ENV CGO_ENABLED=0

RUN go build -ldflags="-s -w" -o /out/svarog ./cmd
RUN go build -ldflags="-s -w" -o /out/mcp-stdio ./cmd/mcp-stdio
RUN go build -ldflags="-s -w" -o /out/mcp-http ./cmd/mcp-http

# -----------------------------------------------------------------------------
# Application image (gRPC + HTTP gateway)
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS app

COPY --from=builder /out/svarog /svarog

EXPOSE 8080 9090

ENTRYPOINT ["/svarog"]

# -----------------------------------------------------------------------------
# MCP stdio image (for sidecar / local exec; not used by default compose)
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS mcp-stdio

COPY --from=builder /out/mcp-stdio /mcp-stdio

ENTRYPOINT ["/mcp-stdio"]

# -----------------------------------------------------------------------------
# MCP streamable-HTTP image
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS mcp-http

COPY --from=builder /out/mcp-http /mcp-http

EXPOSE 8000

ENTRYPOINT ["/mcp-http"]
