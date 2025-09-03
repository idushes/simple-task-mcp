# Simple Task MCP Server

MCP (Model Context Protocol) server for task management, designed for use by AI agents.

## Features

- **User Management**: Create and manage users with admin privileges
- **JWT Authentication**: Secure API access with JWT tokens via Authorization header
- **PostgreSQL Database**: Persistent storage with automatic migrations
- **Dual Transport Support**: HTTP/SSE (default) and stdio
- **CORS Support**: For cross-origin requests in web applications
- **Graceful Shutdown**: Proper cleanup on server termination

## Architecture

The server uses PostgreSQL for data persistence and JWT tokens for authentication. Admin users can create new users, and in future releases, users will be able to manage tasks.

### Database Schema

**Users Table**:
- `id` (UUID) - Primary key
- `name` (VARCHAR) - User name
- `is_admin` (BOOLEAN) - Admin privileges
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

**Tasks Table** (prepared for future use):
- `id` (UUID) - Primary key
- `description` (TEXT) - Task description
- `status` (VARCHAR) - Task status (pending, in_progress, waiting_for_user, completed, cancelled)
- `created_by` (UUID) - Reference to user
- `assigned_to` (UUID) - Reference to user
- `is_archived` (BOOLEAN)
- Timestamps for creation, update, completion, and archiving

## Setup

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Git

### Installation

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

## Available Tools

### create_user

Creates a new user in the system. Only admins can use this tool.

**Parameters**:
- `name` (string, required) - Name of the user
- `is_admin` (boolean, optional) - Whether the user should have admin privileges (default: false)

**Authentication**: Required (JWT token in Authorization header)

**Example Response**:
```json
{
  "success": true,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "is_admin": false,
    "token": "eyJhbGciOiJIUzI1NiIs..."
  },
  "message": "User 'John Doe' created successfully"
}
```

## Development

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

### Project Structure

```
simple-task-mcp/
├── auth/               # JWT authentication
├── cmd/               # Command line tools
│   └── create-admin/  # Admin user creation utility
├── config/            # Configuration management
├── database/          # Database connection and migrations
├── models/            # Data models
├── server/            # HTTP middleware
├── tools/             # MCP tool implementations
└── main.go            # Application entry point
```

## Roadmap

- [x] Basic MCP server setup
- [x] PostgreSQL integration
- [x] JWT authentication
- [x] User management (create_user tool)
- [ ] Task creation (create_task tool)
- [ ] Task retrieval (get_next_task tool)
- [ ] Task updates (update_task tool)
- [ ] Task status management (update_task_status tool)
- [ ] Task archiving (archive_task tool)
- [ ] Docker containerization
- [ ] GitHub Actions CI/CD

## Security

- JWT tokens are used for authentication
- Tokens are passed via standard Authorization header
- Admin privileges are required for user management
- Database connections use prepared statements to prevent SQL injection

## License

MIT License - see LICENSE file for details