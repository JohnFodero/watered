# Task 6: Docker Configuration

## Objective
Create Docker configuration for development and production deployment.

## Requirements
- [ ] Create multi-stage Dockerfile for Go application
- [ ] Set up docker-compose.yml for development
- [ ] Configure environment variable management
- [ ] Add database persistence with Docker volumes
- [ ] Set up health checks and logging
- [ ] Create development vs production configurations

## Docker Files
- [ ] `Dockerfile` - Multi-stage build for Go app
- [ ] `docker-compose.yml` - Development environment
- [ ] `docker-compose.prod.yml` - Production configuration
- [ ] `.dockerignore` - Exclude unnecessary files

## Development Setup
- [ ] Hot reload for Go development (optional)
- [ ] Volume mounts for local development
- [ ] Environment variable configuration
- [ ] Database initialization and migrations
- [ ] Port mapping and service discovery

## Production Considerations
- [ ] Minimal container size
- [ ] Non-root user execution
- [ ] Health check endpoints
- [ ] Logging configuration
- [ ] Security hardening

## Environment Variables
```
# Application
PORT=8080
SESSION_SECRET=your-secret-key

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret

# Database
DATABASE_PATH=/data/watered.db

# Admin
ADMIN_EMAILS=admin@example.com
ALLOWED_EMAILS=user1@example.com,user2@example.com
```

## Docker Compose Services
- `app` - Go web application
- `db` - SQLite database (file-based)
- Optional: `nginx` - Reverse proxy for production

## Success Criteria
- Application runs in Docker containers
- Development environment starts with single command
- Database data persists between container restarts
- Environment variables are properly configured
- Health checks work correctly
- Logs are accessible and well-formatted

## Next Steps
- Task 7: Testing and deployment
- Task 8: Documentation and monitoring