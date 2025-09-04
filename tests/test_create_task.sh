#!/bin/bash

# Script to test create_task functionality

echo "=== Testing create_task tool ==="
echo ""

# Check if token is provided as argument
if [ -z "$1" ]; then
    echo "Usage: ./test_create_task.sh <JWT_TOKEN> <USERNAME_TO_ASSIGN>"
    echo ""
    echo "First, you need to get a JWT token by creating a user."
    echo "Run the following to create an admin user:"
    echo "  ../create-admin"
    echo ""
    echo "Then use the token from the output to test create_task."
    echo "Example: ./test_create_task.sh <TOKEN> 'John Doe'"
    exit 1
fi

JWT_TOKEN=$1
ASSIGN_TO_USERNAME=${2:-"Admin User"}

# Create a test task
echo "Creating a test task..."
echo ""

curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "create_task",
      "arguments": {
        "description": "Test task: Implement user authentication",
        "assigned_to": "'"$ASSIGN_TO_USERNAME"'"
      }
    },
    "id": 1
  }' | jq .

echo ""
echo "=== Test completed ==="

# Usage example in comment
: '
Example usage:
1. First create an admin user to get a token:
   cd .. && ./create-admin

2. Copy the token from the output

3. Run this test:
   ./test_create_task.sh <TOKEN> "Admin User"

Or assign to a different user:
   ./test_create_task.sh <TOKEN> "John Doe"
'
