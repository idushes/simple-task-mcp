# Simple Task MCP Server

MCP (Model Context Protocol) server for task management, designed for use by AI agents.

## Stage 1: Basic Infrastructure

This is the first stage of development with basic MCP server setup and protocol-level ping support.

### Features

- Dual transport support: HTTP/SSE (default) and stdio
- HTTP endpoint at `/mcp` for web-based clients
- CORS support for cross-origin requests
- Stateless mode for simplified HTTP client integration
- Configuration loading from environment variables
- Built-in ping support at protocol level

### Setup

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the server:
   ```bash
   go build -o simple-task-mcp
   ```

### Running

The server supports two transport modes:

#### HTTP Transport (Default)
```bash
./simple-task-mcp
# or explicitly:
./simple-task-mcp --transport http
```

The HTTP server will start on the configured port (default: 8080) with the endpoint at:
```
http://localhost:8080/mcp
```

The HTTP mode runs in stateless mode and includes CORS support for web-based clients.

#### Stdio Transport
```bash
./simple-task-mcp --transport stdio
```

This mode is used by MCP Inspector and Claude Desktop.

### Configuration

Configuration is loaded from environment variables:

- `MCP_SERVER_PORT` - HTTP server port for SSE endpoint (default: 8080)
- `LOG_LEVEL` - Logging level: debug, info, warn, error (default: info)

### Protocol Features

#### Ping
The server supports the MCP ping protocol for connection health checks. Send a ping request:

The server responds with an empty result `{}` as per MCP specification.

### MCP Client Configuration

#### For MCP Inspector or Claude Desktop (stdio):
```json
{
  "mcpServers": {
    "simple-task-mcp": {
      "command": "/path/to/simple-task-mcp"
    }
  }
}
```

#### For web-based clients (HTTP/SSE):
```json
{
  "mcpServers": {
    "simple-task-mcp": {
      "url": "http://localhost:8080/mcp",
      "transport": "sse"
    }
  }
}
```

### Development

For development, you can run directly with Go:
```bash
go run main.go
```

### Testing

To test the server functionality, you can use the included test script:

```bash
./test_connection.sh
```

Or manually test with curl:

```bash
# For HTTP transport:
# Initialize connection
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }'

# Send ping request
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "ping"
  }'
```

### Server Configuration

- **Transport Modes**:
  - **http**: HTTP with Server-Sent Events (SSE) for web clients (default)
  - **stdio**: Standard input/output for desktop clients
- **Stateless Mode**: HTTP transport operates in stateless mode
- **CORS**: Enabled for HTTP transport to support cross-origin requests
