# MCP Simulator

Simulates N virtual MCP servers on a single port using the [Streamable HTTP](https://modelcontextprotocol.io/specification/2025-03-26/basic/transports#streamable-http) transport. Each server maintains a random set of tools and periodically mutates them, notifying connected clients via SSE.

Built for load-testing an MCP client that holds persistent connections to thousands of servers.

## Quick Start

```bash
# Run with defaults (1000 servers, port 8080)
go run ./cmd/simulator/

# Custom configuration
go run ./cmd/simulator/ \
  --servers 1000 \
  --port 9090 \
  --min-tools 3 \
  --max-tools 10 \
  --min-mutation-interval 5s \
  --max-mutation-interval 60s

# Build and run binary
go build -o mcp-simulator ./cmd/simulator/
./mcp-simulator --servers 10000 --port 9090
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | HTTP server port |
| `--servers` | `1000` | Number of virtual MCP servers |
| `--min-tools` | `3` | Minimum tools per server at init |
| `--max-tools` | `10` | Maximum tools per server at init |
| `--min-mutation-interval` | `5s` | Minimum time between tool mutations |
| `--max-mutation-interval` | `60s` | Maximum time between tool mutations |

## API

Each virtual server exposes a single Streamable HTTP endpoint:

```
POST /server/{id}/mcp
```

### Supported JSON-RPC methods

| Method | Description | Response |
|--------|-------------|----------|
| `initialize` | MCP handshake | Server capabilities |
| `notifications/initialized` | Client ready signal | 202 Accepted |
| `tools/list` | List current tools | Array of tools with inputSchema |

### Regular JSON response

```bash
curl -X POST http://localhost:9090/server/0/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}'
```

### SSE stream with notifications

Add `Accept: text/event-stream` to receive the response as an SSE event, followed by `notifications/tools/list_changed` events whenever the server's tools change:

```bash
curl -N -X POST http://localhost:9090/server/0/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: text/event-stream" \
  -d '{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}'
```

## Architecture

```
cmd/simulator/main.go        Entry point, CLI flags, graceful shutdown
internal/jsonrpc/jsonrpc.go   JSON-RPC 2.0 request/response/notification structs
internal/tools/tools.go       Random tool generation and mutation
internal/server/server.go     VirtualServer state, pub/sub, mutation loop, Manager
internal/server/handler.go    HTTP handler, JSON-RPC dispatch, SSE streaming
```

- One goroutine per server runs a mutation loop (add/remove/update a random tool at random intervals)
- Mutations broadcast `notifications/tools/list_changed` to all SSE subscribers
- Zero external dependencies — standard library only
