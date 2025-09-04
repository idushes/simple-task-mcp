#!/bin/bash

# Test MCP server connection

echo "Testing MCP server initialization..."

# Send initialize request
curl -X POST http://localhost:6688/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {
        "tools": {}
      },
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }' | jq .

echo -e "\n\nTesting protocol ping..."

# Send ping request
curl -X POST http://localhost:6688/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "ping"
  }' | jq .

echo -e "\n\nListing available tools..."

# List tools (should be empty for now)
curl -X POST http://localhost:6688/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/list",
    "params": {}
  }' | jq .
