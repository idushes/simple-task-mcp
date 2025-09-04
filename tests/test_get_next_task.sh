#!/bin/bash

# Test script for get_next_task tool

# Configuration
SERVER_URL="http://localhost:8080/mcp"
ADMIN_JWT="YOUR_ADMIN_JWT_TOKEN"
USER_JWT="YOUR_USER_JWT_TOKEN"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to make MCP request
make_request() {
    local jwt_token=$1
    local request_data=$2
    
    curl -X POST "$SERVER_URL" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $jwt_token" \
        -d "$request_data" \
        -s
}

echo -e "${YELLOW}Starting get_next_task tool tests...${NC}"

# Test 1: Get tasks with pending status
echo -e "\n${YELLOW}Test 1: Get task with pending status${NC}"
REQUEST_DATA='{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
        "name": "get_next_task",
        "arguments": {
            "statuses": ["pending"]
        }
    },
    "id": 1
}'

RESPONSE=$(make_request "$USER_JWT" "$REQUEST_DATA")
echo "Response: $RESPONSE"

if echo "$RESPONSE" | grep -q '"error"'; then
    echo -e "${RED}✗ Test 1 failed${NC}"
else
    echo -e "${GREEN}✓ Test 1 passed${NC}"
fi

# Test 2: Get tasks with multiple statuses
echo -e "\n${YELLOW}Test 2: Get task with multiple statuses${NC}"
REQUEST_DATA='{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
        "name": "get_next_task",
        "arguments": {
            "statuses": ["pending", "in_progress", "waiting_for_user"]
        }
    },
    "id": 2
}'

RESPONSE=$(make_request "$USER_JWT" "$REQUEST_DATA")
echo "Response: $RESPONSE"

if echo "$RESPONSE" | grep -q '"error"'; then
    echo -e "${RED}✗ Test 2 failed${NC}"
else
    echo -e "${GREEN}✓ Test 2 passed${NC}"
fi

# Test 3: Get completed tasks
echo -e "\n${YELLOW}Test 3: Get completed tasks${NC}"
REQUEST_DATA='{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
        "name": "get_next_task",
        "arguments": {
            "statuses": ["completed"]
        }
    },
    "id": 3
}'

RESPONSE=$(make_request "$USER_JWT" "$REQUEST_DATA")
echo "Response: $RESPONSE"

if echo "$RESPONSE" | grep -q '"error"'; then
    echo -e "${RED}✗ Test 3 failed${NC}"
else
    echo -e "${GREEN}✓ Test 3 passed${NC}"
fi

# Test 4: Invalid status
echo -e "\n${YELLOW}Test 4: Invalid status (should fail)${NC}"
REQUEST_DATA='{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
        "name": "get_next_task",
        "arguments": {
            "statuses": ["invalid_status"]
        }
    },
    "id": 4
}'

RESPONSE=$(make_request "$USER_JWT" "$REQUEST_DATA")
echo "Response: $RESPONSE"

if echo "$RESPONSE" | grep -q '"error"' && echo "$RESPONSE" | grep -q 'invalid status'; then
    echo -e "${GREEN}✓ Test 4 passed (correctly rejected invalid status)${NC}"
else
    echo -e "${RED}✗ Test 4 failed (should reject invalid status)${NC}"
fi

# Test 5: Empty statuses array (should use default)
echo -e "\n${YELLOW}Test 5: Empty statuses array (should use default [\"pending\"])${NC}"
REQUEST_DATA='{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
        "name": "get_next_task",
        "arguments": {
            "statuses": []
        }
    },
    "id": 5
}'

RESPONSE=$(make_request "$USER_JWT" "$REQUEST_DATA")
echo "Response: $RESPONSE"

if echo "$RESPONSE" | grep -q '"error"'; then
    echo -e "${RED}✗ Test 5 failed (should use default value)${NC}"
else
    echo -e "${GREEN}✓ Test 5 passed (uses default value [\"pending\"])${NC}"
fi

# Test 5.1: No statuses parameter (should use default)
echo -e "\n${YELLOW}Test 5.1: No statuses parameter (should use default [\"pending\"])${NC}"
REQUEST_DATA='{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
        "name": "get_next_task",
        "arguments": {}
    },
    "id": 51
}'

RESPONSE=$(make_request "$USER_JWT" "$REQUEST_DATA")
echo "Response: $RESPONSE"

if echo "$RESPONSE" | grep -q '"error"'; then
    echo -e "${RED}✗ Test 5.1 failed (should use default value)${NC}"
else
    echo -e "${GREEN}✓ Test 5.1 passed (uses default value [\"pending\"])${NC}"
fi

# Test 6: No authorization
echo -e "\n${YELLOW}Test 6: No authorization (should fail)${NC}"
REQUEST_DATA='{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
        "name": "get_next_task",
        "arguments": {
            "statuses": ["pending"]
        }
    },
    "id": 6
}'

RESPONSE=$(curl -X POST "$SERVER_URL" \
    -H "Content-Type: application/json" \
    -d "$REQUEST_DATA" \
    -s)
echo "Response: $RESPONSE"

if echo "$RESPONSE" | grep -q '"error"' && echo "$RESPONSE" | grep -qi 'authorization'; then
    echo -e "${GREEN}✓ Test 6 passed (correctly requires authorization)${NC}"
else
    echo -e "${RED}✗ Test 6 failed (should require authorization)${NC}"
fi

echo -e "\n${YELLOW}Tests completed!${NC}"
echo -e "${YELLOW}Remember to:${NC}"
echo -e "1. Start the MCP server: ${GREEN}go run main.go${NC}"
echo -e "2. Update JWT tokens in this script with valid tokens"
echo -e "3. Create some test tasks before running this script"
echo -e "4. Make this script executable: ${GREEN}chmod +x test_get_next_task.sh${NC}"
