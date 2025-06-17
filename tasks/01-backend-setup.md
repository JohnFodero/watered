# Task 1: Backend Setup

## Objective
Initialize the Go backend with proper project structure and dependencies.

## Requirements
- [ ] Initialize `go.mod` with module name
- [ ] Create basic project directory structure (`cmd/`, `internal/`, `pkg/`)
- [ ] Set up main server entry point in `cmd/server/main.go`
- [ ] Add initial HTTP server with basic health check endpoint
- [ ] Create placeholder packages for auth, handlers, models, storage
- [ ] Add basic logging setup
- [ ] Write initial tests for HTTP server

## Dependencies to Add
- `github.com/go-chi/chi/v5` - Lightweight HTTP router
- `golang.org/x/oauth2` - OAuth2 client (for future Google SSO)
- Standard library packages for HTTP, JSON, logging

## Success Criteria
- Go server starts and responds to health check
- Project structure follows Go best practices
- Tests pass with `go test ./...`
- Documentation explains how to run the server

## Next Steps
- Task 2: Frontend structure setup
- Task 3: Google SSO integration