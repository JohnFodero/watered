# Task 4: Plant State Management API

## Objective
Create API endpoints for managing plant watering state and timer functionality.

## Requirements
- [ ] Design plant state data model
- [ ] Implement in-memory storage for plant state
- [ ] Create API endpoints for plant operations
- [ ] Add timer calculation logic
- [ ] Implement plant status determination (healthy/withered)
- [ ] Add plant state persistence (SQLite)
- [ ] Create comprehensive tests for all endpoints

## Data Model
```go
type PlantState struct {
    ID           int       `json:"id"`
    Name         string    `json:"name"`
    LastWatered  time.Time `json:"last_watered"`
    TimeoutHours int       `json:"timeout_hours"`
    WateredBy    string    `json:"watered_by"` // user email
}
```

## API Endpoints
- `GET /api/plant` - Get current plant state
- `POST /api/plant/water` - Record plant watering (resets timer)
- `GET /api/plant/status` - Get plant health status (healthy/withered)
- `GET /api/plant/timer` - Get time since last watering

## Business Logic
- [ ] Calculate time since last watering
- [ ] Determine plant health based on timeout
- [ ] Record who watered the plant last
- [ ] Handle timezone considerations
- [ ] Validate watering actions

## Storage Layer
- [ ] Interface for storage operations
- [ ] In-memory implementation for development
- [ ] SQLite implementation for persistence
- [ ] Migration system for database schema

## Success Criteria
- Plant state persists between server restarts
- API correctly calculates plant health status
- Timer functionality works accurately
- All endpoints have comprehensive test coverage
- Frontend can interact with all API endpoints

## Next Steps
- Task 5: Admin panel functionality
- Task 6: Docker configuration