# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gonsole is a WebSocket-based remote console system for Go applications. It provides a web-based interface for monitoring system status, adjusting parameters, and executing commands in real-time, similar to system consoles found in Linux, MySQL, and other sophisticated software systems.

## Development Commands

### Frontend (web/)
```bash
cd web/
yarn install              # Install dependencies
yarn dev                  # Start dev server on port 3001
yarn build                # Build for production (outputs to web/dist/)
```

### Backend (Go)
```bash
# Run demo server
go run examples/demo.go   # Starts on port 8888, access at https://localhost:8888/console

# Run tests
go test ./...

# Build with version info
FLAGS="-w -s -X $IMPORT_PATH.GitBranchName=`git rev-parse --abbrev-ref HEAD` ..."
go build -ldflags "$FLAGS" -mod vendor
```

## Architecture

### Client-Server Model
The system uses WebSocket communication between a SolidJS frontend and Go backend:

```
Web Browser (SolidJS UI) <--WebSocket--> Go Server (gonsole)
```

### Core Components

1. **Console Layer** (`console.go`, `console_service.go`)
   - Main console server implementation
   - Manages commands and topics
   - Handles HTTP/WebSocket routing
   - Integrates pprof for profiling

2. **Road Networking Layer** (`road/`)
   - Custom WebSocket framework adapted from Pitaya v1.1.1
   - Session management and message routing
   - Multiple serialization formats (JSON primary)
   - Connection pooling and epoll-based I/O
   - Built-in scan defender for security

3. **Authentication** (`jwtx/`, `beans/command_auth.go`)
   - JWT-based authentication
   - Username/password verification
   - Token-based auto-login (configurable time window)

4. **Web Frontend** (`web/`)
   - SolidJS-based reactive UI
   - WebSocket client implementation
   - Command history and auto-completion

### Architectural Patterns

- **Command Pattern**: Commands are registered with metadata (name, access control, handler)
- **Pub/Sub for Topics**: Periodic data broadcasting to subscribed clients
- **Functional Options**: Configuration via `With*()` option functions
- **Component Registration**: Services auto-discover methods and register routes

## Registration Patterns

### Commands
```go
console.RegisterCommand(&gonsole.Command{
    Name: "mycommand",
    Note: "Description",
    Flag: gonsole.FlagPublic,  // or 0 for private commands
    Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
        return gonsole.NewDefaultResponse(data), nil
    },
})
```

### Topics
```go
console.RegisterTopic(&gonsole.Topic{
    Name: "metrics",
    Interval: 10 * time.Second,
    BuildResponse: func() *gonsole.Response {
        return gonsole.NewDefaultResponse(getMetrics())
    },
})
```

### Services
```go
console.RegisterService("myservice", newMyService())
// Methods are auto-registered as routes like "myservice.methodName"
```

## Response Types

- `NewDefaultResponse(data)` - JSON data response
- `NewHtmlResponse(html)` - HTML content
- `NewTableResponse(data)` - Structured table data with sorting
- `ToHtmlTable(data)` - Convert struct/slice to HTML table

## Route Naming Convention

Routes follow `<service>.<method>` format. CamelCase Go methods are automatically converted to snake_case routes via `ToSnakeName()`.

## Configuration Options

All configuration via functional options:

```go
gonsole.NewConsole(mux,
    gonsole.WithPort(8888),
    gonsole.WithPageTemplate("web/dist/console.html"),
    gonsole.WithUserPasswords(map[string]string{"admin": "secret"}),
    gonsole.WithEnablePProf(true),
    gonsole.WithTls(true),
    gonsole.WithDirectory("ws"),  // URL path prefix
)
```

## Command Flags

```go
const (
    flagBuiltin   = 0x0001  // Built-in commands
    FlagPublic    = 0x0002  // No auth required
    FlagInvisible = 0x0004  // Hidden from help
)
```

## Special Features

### Road SendStream
For progressive/streaming responses:
```go
road.SendStream(session, "data chunk", false)  // false = not final
road.SendStream(session, "", true)  // true = final
```

### Session Echo
Atomic session operations:
```go
session.Echo(func() {
    // Atomic operation on session
})
```

## Built-in Commands

- `help` - List all commands
- `auth <username>` - Authenticate
- `log.list` - List log files
- `head <file>`, `tail <file>` - View file contents
- `history` - Command history
- `top` - Process statistics
- `deadlock.detect` - Detect deadlocks

## SSL/TLS

Self-signed certificates in `res/ssl/` (generated with mkcert). HTTPS with TLS is enabled by default.

## Security

- Scan defender limits connections per IP (max 5)
- JWT-based authentication with configurable auto-login window
- Public/private command distinction
- pprof requires recent authentication (10 min window)

## Project Structure Notes

- Frontend assets are embedded into Go binary via `init.go` force include
- Uses vendored dependencies
- Documentation and README in Chinese, code in English
- Go version 1.22+ required
