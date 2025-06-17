# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

"Watered" is a simple web application for tracking plant watering between partners. Users log in via Google SSO to see a visual plant representation that changes based on watering status. Tapping the plant resets the watering timer. Includes an admin panel for managing timeout settings and authorized email addresses.

**Tech Stack:**
- Frontend: Alpine.js, HTML/CSS/JavaScript
- Backend: Go (Golang)
- Authentication: Google SSO
- Deployment: Docker (self-hosted initially)
- Future: Nix shells, cloud hosting

## Development Guidelines

**Core Principles:**
- Keep it simple first, add complexity feature by feature
- Test everything - write tests for all functionality
- Document setup and testing procedures thoroughly
- Focus on learning frontend development and Go skills

**Coding Conventions:**
- Write comprehensive tests for all new features
- Follow Go best practices and idiomatic patterns
- Use Alpine.js for reactive frontend behavior
- Maintain clean, readable code structure

## Development Setup

### Prerequisites
- Go 1.21+
- Docker and Docker Compose
- Node.js (for any frontend tooling)

### Running the Application
```bash
# Development with Docker
docker-compose up --build

# Run Go backend directly
go run cmd/server/main.go

# Run tests
go test ./...

# Build for production
go build -o bin/watered cmd/server/main.go
```

### Project Structure
```
watered/
├── cmd/server/          # Application entrypoint
├── internal/           # Private application code
│   ├── auth/          # Authentication logic
│   ├── handlers/      # HTTP handlers
│   ├── models/        # Data models
│   └── storage/       # Database layer
├── pkg/               # Public libraries
├── web/               # Frontend assets
│   ├── static/        # CSS, JS, images
│   └── templates/     # HTML templates
├── tasks/             # Development task files
└── docker-compose.yml # Development environment
```

## Architecture

**Frontend:**
- Single-page application using Alpine.js for reactivity
- Static HTML templates served by Go backend
- Plant visualization with CSS animations/transitions
- Mobile-responsive design

**Backend:**
- Go HTTP server with Chi router (github.com/go-chi/chi/v5)
- RESTful API for plant state management
- Google OAuth2 integration for authentication
- In-memory storage initially, SQLite for persistence
- Admin API endpoints for configuration

**Authentication Flow:**
1. User visits app, redirected to Google OAuth
2. Google returns with user info and token
3. Backend validates token and checks email whitelist
4. Session established with secure cookies
5. Frontend receives authentication state

**Data Models:**
- User (email, name, admin status)
- PlantState (last_watered, timeout_hours)
- AdminConfig (timeout, whitelisted_emails)

## Design Theme

**Color Palette (Monokai Pro Light Filter Sun):**
- Primary Background: `#f8efe7` (soft warm beige)
- Secondary Background: `#ede5de` (slightly darker warm neutral)
- Activity Background: `#ded5d0` (muted rose-taupe)
- Primary Text: `#2c232d` (deep plum-brown)
- Accent Color: `#cd4770` (vibrant rose-pink)
- Muted Text: `#91898a` (soft gray-mauve)

**Design Principles:**
- Soft, warm color scheme with delicate neutrals
- Gentle, nurturing interface that feels organic
- Subtle animations for plant state transitions
- Clean, minimalist design focused on usability