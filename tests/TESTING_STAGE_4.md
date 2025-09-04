# Testing Stage 4: get_next_task Tool

## Overview
This document describes how to test the `get_next_task` tool functionality.

## Prerequisites
1. MCP server is running (`go run main.go`)
2. Database is set up with users and tasks
3. You have valid JWT tokens for testing

## Setup Test Data

### 1. Create Test Users
Use the scripts from Stage 3 to create test users if not already done.

### 2. Create Test Tasks
Run the test_create_task.sh script to create several tasks with different statuses:

```bash
# Create tasks with different statuses
./tests/test_create_task.sh
```

### 3. Update Task Statuses (Manual)
Currently, you'll need to manually update some task statuses in the database for testing:

```sql
-- Connect to your database
psql -U your_user -d your_database

-- Update some tasks to different statuses
UPDATE tasks SET status = 'in_progress' WHERE id = 'some-task-id';
UPDATE tasks SET status = 'waiting_for_user' WHERE id = 'another-task-id';
UPDATE tasks SET status = 'completed', completed_at = NOW() WHERE id = 'third-task-id';
```

## Running Tests

### 1. Update JWT Tokens
Edit `test_get_next_task.sh` and update the JWT tokens:
```bash
ADMIN_JWT="your-admin-jwt-token"
USER_JWT="your-user-jwt-token"
```

### 2. Run the Test Script
```bash
./tests/test_get_next_task.sh
```

## Test Cases

### Test 1: Get Pending Tasks
- Filters tasks by 'pending' status
- Returns the oldest pending task where user is creator or assignee

### Test 2: Multiple Status Filter
- Filters by multiple statuses: ["pending", "in_progress", "waiting_for_user"]
- Returns the oldest matching task

### Test 3: Get Completed Tasks
- Filters by 'completed' status
- Should include completed_at timestamp

### Test 4: Invalid Status
- Tests with invalid status value
- Should return error: "invalid status: {status}"

### Test 5: Empty Status Array
- Tests with empty statuses array
- Should use default value ["pending"]
- Should not return error

### Test 5.1: No Status Parameter
- Tests without statuses parameter at all
- Should use default value ["pending"]
- Should not return error

### Test 6: No Authorization
- Tests without JWT token
- Should return authorization error

## Expected Response Format

### Success Response (Task Found):
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\n  \"id\": \"uuid-here\",\n  \"description\": \"Task description\",\n  \"status\": \"pending\",\n  \"created_by\": \"creator_username\",\n  \"created_by_id\": \"creator-uuid\",\n  \"assigned_to\": \"assignee_username\",\n  \"assigned_to_id\": \"assignee-uuid\",\n  \"created_at\": \"2024-01-20T10:00:00Z\",\n  \"updated_at\": \"2024-01-20T10:00:00Z\"\n}"
      }
    ]
  },
  "id": 1
}
```

### Success Response (No Tasks):
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "null"
      }
    ]
  },
  "id": 1
}
```

### Error Response:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Error message here"
      }
    ],
    "isError": true
  },
  "id": 1
}
```

## Manual Testing with curl

### Example 1: Get pending tasks (explicit)
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
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
```

### Example 2: Use default status (pending)
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "get_next_task",
      "arguments": {}
    },
    "id": 1
  }'
```

## Verification Checklist

- [ ] Returns only non-archived tasks
- [ ] Filters correctly by provided statuses
- [ ] Returns tasks where current user is creator OR assignee
- [ ] Returns null when no matching tasks found
- [ ] Orders tasks by created_at (oldest first)
- [ ] Returns only one task
- [ ] Includes both user names and UUIDs in response
- [ ] Validates all status values
- [ ] Requires valid JWT authentication
- [ ] Handles errors gracefully

## Notes

- The tool returns the oldest matching task (by created_at)
- Tasks are filtered by the authenticated user (must be creator or assignee)
- Archived tasks are never returned
- The response includes both human-readable usernames and UUIDs for flexibility
