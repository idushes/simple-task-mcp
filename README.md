# Simple Task MCP Server

MCP (Model Context Protocol) server for task management, designed for use by AI agents.

## Features

- **User Management**: Create and manage users with admin privileges
- **Task Management**: Complete task lifecycle management (create, get, complete, cancel, comment)
- **JWT Authentication**: Secure API access with JWT tokens via Authorization header
- **PostgreSQL Database**: Persistent storage with automatic migrations
- **Dual Transport Support**: HTTP/SSE (default) and stdio
- **CORS Support**: For cross-origin requests in web applications
- **Graceful Shutdown**: Proper cleanup on server termination

## Architecture

The server uses PostgreSQL for data persistence and JWT tokens for authentication. Admin users can create and manage users, while all authenticated users can manage tasks through the complete task lifecycle.

### Database Schema

**Users Table**:
- `id` (UUID) - Primary key
- `name` (VARCHAR) - User name (unique)
- `description` (TEXT) - Optional user description
- `is_admin` (BOOLEAN) - Admin privileges
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

**Tasks Table**:
- `id` (UUID) - Primary key
- `description` (TEXT) - Task description
- `status` (VARCHAR) - Task status (pending, in_progress, waiting_for_user, completed, cancelled)
- `created_by` (UUID) - Reference to user
- `assigned_to` (UUID) - Reference to user
- `result` (TEXT) - Task result or cancellation reason
- `is_archived` (BOOLEAN)
- Timestamps for creation, update, completion, and archiving

**Task Comments Table**:
- `id` (UUID) - Primary key
- `task_id` (UUID) - Reference to task
- `created_by` (UUID) - Reference to user
- `comment` (TEXT) - Comment text
- `created_at` (TIMESTAMP)

## Setup

### Prerequisites

- Go 1.24 or higher (for local development)
- PostgreSQL 12 or higher (for local development)
- Git
- Docker and Docker Compose (for containerized setup)

### Installation

#### Option 1: Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/dushes/simple-task-mcp.git
   cd simple-task-mcp
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create PostgreSQL database:
   ```sql
   CREATE DATABASE task_manager;
   ```

4. Configure environment:
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

5. Build the project:
   ```bash
   go build -o simple-task-mcp .
   go build -o create-admin ./cmd/create-admin
   ```

#### Option 2: Docker Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/dushes/simple-task-mcp.git
   cd simple-task-mcp
   ```

2. Build the Docker image:
   ```bash
   docker build -t simple-task-mcp .
   ```

3. Run the container:
   ```bash
   docker run -p 8080:8080 \
     -e DATABASE_URL=postgres://user:password@host:5432/task_manager \
     -e JWT_SECRET=your-secret-key \
     -e MCP_SERVER_PORT=8080 \
     -e LOG_LEVEL=info \
     simple-task-mcp
   ```

4. The server is now available at http://localhost:8080/mcp

#### Option 3: Using Pre-built Docker Image

You can also use the pre-built Docker image from Docker Hub:

```bash
docker pull dushes/simple-task-mcp:latest

# Run the container
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://user:password@host:5432/task_manager \
  -e JWT_SECRET=your-secret-key \
  -e MCP_SERVER_PORT=8080 \
  -e LOG_LEVEL=info \
  dushes/simple-task-mcp:latest
```

#### Creating Admin User in Docker Container

To create an admin user in a running Docker container:

```bash
# Execute the create-admin utility in the container
docker exec -it <container_id> ./create-admin
```

Alternatively, you can run a one-off command:

```bash
docker run --rm \
  -e DATABASE_URL=postgres://user:password@host:5432/task_manager \
  -e JWT_SECRET=your-secret-key \
  dushes/simple-task-mcp:latest ./create-admin
```

## Configuration

Environment variables (in `.env` file):

```env
# Database
DATABASE_URL=postgres://user:password@localhost:5432/task_manager?sslmode=disable

# MCP Server
MCP_SERVER_PORT=8080

# JWT
JWT_SECRET=your-secret-key-here

# Logging
LOG_LEVEL=info
```

## Usage

### 1. Create Initial Admin User

Before using the server, create an admin user:

```bash
./create-admin
```

This will:
- Connect to the database
- Run migrations
- Create an admin user
- Output the admin's JWT token

Save the JWT token - you'll need it to authenticate API requests.

### 2. Start the Server

#### HTTP Transport (Default)
```bash
./simple-task-mcp
# Server starts at http://localhost:8080/mcp
```

#### Stdio Transport
```bash
./simple-task-mcp --transport stdio
```

### 3. Configure MCP Client

#### For MCP Inspector (HTTP):

1. Transport Type: `Streamable HTTP`
2. URL: `http://localhost:8080/mcp`
3. Authentication:
   - Header Name: `Authorization`
   - Bearer Token: `<your-jwt-token>`

#### For Claude Desktop (HTTP):
```json
{
  "mcpServers": {
    "simple-task-mcp": {
      "serverUrl": "http://simple-task-mcp-service.agents.svc.cluster.local/mcp",
      "requestHeaders": {
        "Authorization": "Bearer <your-jwt-token>"
      }
    }
  }
}
```

#### For Claude Desktop (stdio):
```json
{
  "mcpServers": {
    "simple-task-mcp": {
      "command": "/path/to/simple-task-mcp",
      "args": ["--transport", "stdio"]
    }
  }
}
```

#### For Cursor IDE:
```json
{
  "mcpServers": {
    "simple-task-mcp": {
      "serverUrl": "http://localhost:8080/mcp",
      "requestHeaders": {
        "Authorization": "Bearer <your-jwt-token>"
      }
    }
  }
}
```

## Available Tools

### create_user (Admin Only)
Creates a new user in the system.
- **Parameters**: `name` (required), `is_admin` (optional)
- **Returns**: User details with JWT token

### list_users
Lists users in the system.
- **Parameters**: `limit` (optional - number, default: 100, max: 1000)
- **Returns**: Array of users with their details, count, and limit info

### create_task
Creates a new task and assigns it to a user.
- **Parameters**: `description` (required), `assigned_to` (required - username)
- **Returns**: Task details with creator and assignee names

### get_next_task
Gets the next task for the current user.
- **Parameters**: `statuses` (optional - array, default: ["pending"])
- **Returns**: Single task where user is creator or assignee

### complete_task
Marks a task as completed.
- **Parameters**: `id` (required - task UUID), `result` (optional)
- **Returns**: Updated task details

### cancel_task
Cancels a task with reason.
- **Parameters**: `id` (required - task UUID), `reason` (required)
- **Returns**: Updated task details

### wait_for_user
Sends task to waiting status with comment.
- **Parameters**: `id` (required - task UUID), `comment` (required)
- **Returns**: Updated task details

### generate_token (Admin Only)
Generates new JWT token for existing user.
- **Parameters**: `user_id` (required - user UUID)
- **Returns**: New JWT token and user details

### get_token_info
Gets information about current JWT token.
- **Parameters**: None
- **Returns**: Token details and expiration info

## Development

### CI/CD Pipeline

This project uses GitHub Actions for continuous integration and delivery:

- Automatic Docker image builds on pushes to the `main` branch
- Automatic versioned Docker image builds when tags starting with `v` are pushed
- Images are published to Docker Hub

To use the CI/CD pipeline:

1. Fork or clone this repository
2. Add the following secrets to your GitHub repository:
   - `DOCKER_USERNAME`: Your Docker Hub username
   - `DOCKER_PASSWORD`: Your Docker Hub access token
3. Push to `main` or create a tag starting with `v` (e.g., `v1.0.0`)

### Running in Development Mode

```bash
# Run with live reload (requires air)
air

# Or run directly
go run main.go
```

### Testing Database Connection

```bash
# Build and run the test utility
go run ./cmd/create-admin
```

### Running Tests

Test scripts are located in the `tests/` directory:

```bash
# Test database connection
./tests/test_connection.sh

# Test create_task tool
./tests/test_create_task.sh <JWT_TOKEN> <USERNAME>
```

See `tests/README.md` for more details.

### Project Structure

```
simple-task-mcp/
├── auth/               # JWT authentication
├── cmd/                # Command line tools
│   └── create-admin/   # Admin user creation utility
├── config/             # Configuration management
├── database/           # Database connection and migrations
│   └── migrations/     # SQL migration files
├── models/             # Data models
├── server/             # HTTP middleware
├── tests/              # Test scripts and documentation
├── tools/              # MCP tool implementations
├── Dockerfile          # Docker build instructions
└── main.go             # Application entry point
```

## Roadmap

- [x] Basic MCP server setup
- [x] PostgreSQL integration
- [x] JWT authentication
- [x] User management (create_user tool)
- [x] User listing (list_users tool)
- [x] Task creation (create_task tool)
- [x] Task retrieval (get_next_task tool)
- [x] Task completion (complete_task tool)
- [x] Task cancellation (cancel_task tool)
- [x] Task comments and user interaction (wait_for_user tool)
- [x] Token generation (generate_token tool)
- [x] Token information (get_token_info tool)
- [x] Docker containerization
- [x] GitHub Actions CI/CD

## Security

- JWT tokens are used for authentication
- Tokens are passed via standard Authorization header
- Admin privileges are required for user management
- Database connections use prepared statements to prevent SQL injection
