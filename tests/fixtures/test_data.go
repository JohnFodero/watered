package fixtures

import (
	"time"
	"watered/internal/models"
)

// TestPlantState creates a test plant state for testing
func TestPlantState() *models.PlantState {
	now := time.Now()
	lastWatered := now.Add(-12 * time.Hour) // 12 hours ago

	return &models.PlantState{
		ID:           1,
		Name:         "Test Plant",
		LastWatered:  &lastWatered,
		TimeoutHours: 24,
		WateredBy:    "test@example.com",
		CreatedAt:    now.Add(-24 * time.Hour),
		UpdatedAt:    now,
	}
}

// TestPlantStateNeverWatered creates a test plant that has never been watered
func TestPlantStateNeverWatered() *models.PlantState {
	now := time.Now()

	return &models.PlantState{
		ID:           2,
		Name:         "Never Watered Plant",
		LastWatered:  nil,
		TimeoutHours: 24,
		WateredBy:    "",
		CreatedAt:    now.Add(-48 * time.Hour),
		UpdatedAt:    now,
	}
}

// TestPlantStateOverdue creates a test plant that is overdue for watering
func TestPlantStateOverdue() *models.PlantState {
	now := time.Now()
	lastWatered := now.Add(-36 * time.Hour) // 36 hours ago, overdue for 24h timeout

	return &models.PlantState{
		ID:           3,
		Name:         "Overdue Plant",
		LastWatered:  &lastWatered,
		TimeoutHours: 24,
		WateredBy:    "test@example.com",
		CreatedAt:    now.Add(-72 * time.Hour),
		UpdatedAt:    now,
	}
}

// TestUser creates a test user
func TestUser() *models.User {
	return &models.User{
		Email:    "test@example.com",
		Name:     "Test User",
		IsAdmin:  false,
		JoinedAt: time.Now().Add(-7 * 24 * time.Hour), // 7 days ago
	}
}

// TestAdminUser creates a test admin user
func TestAdminUser() *models.User {
	return &models.User{
		Email:    "admin@example.com",
		Name:     "Admin User",
		IsAdmin:  true,
		JoinedAt: time.Now().Add(-30 * 24 * time.Hour), // 30 days ago
	}
}

// TestAdminConfig creates a test admin configuration
func TestAdminConfig() *models.AdminConfig {
	return &models.AdminConfig{
		TimeoutHours:  24,
		AllowedEmails: []string{"test@example.com", "user@example.com"},
		AdminEmails:   []string{"admin@example.com"},
		LastModified:  time.Now(),
		ModifiedBy:    "admin@example.com",
	}
}

// TestWateringEvent creates a test watering event
func TestWateringEvent() *models.PlantWateringEvent {
	return &models.PlantWateringEvent{
		ID:        1,
		PlantID:   1,
		WateredAt: time.Now().Add(-6 * time.Hour),
		WateredBy: "test@example.com",
	}
}

// TestJSONPayloads contains common JSON payloads for testing
var TestJSONPayloads = struct {
	ValidPlantSettings   string
	InvalidPlantSettings string
	ValidUserAdd         string
	InvalidUserAdd       string
	ValidTimeoutUpdate   string
	InvalidTimeoutUpdate string
}{
	ValidPlantSettings: `{
		"name": "My Beautiful Plant",
		"timeoutHours": 48
	}`,
	InvalidPlantSettings: `{
		"name": "",
		"timeoutHours": -1
	}`,
	ValidUserAdd: `{
		"email": "newuser@example.com"
	}`,
	InvalidUserAdd: `{
		"email": "invalid-email"
	}`,
	ValidTimeoutUpdate: `{
		"timeoutHours": 72
	}`,
	InvalidTimeoutUpdate: `{
		"timeoutHours": 200
	}`,
}

// TestHTTPHeaders contains common HTTP headers for testing
var TestHTTPHeaders = struct {
	JSON        map[string]string
	FormData    map[string]string
	InvalidAuth map[string]string
}{
	JSON: map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	},
	FormData: map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	},
	InvalidAuth: map[string]string{
		"Authorization": "Bearer invalid-token",
	},
}

// TestScenarios defines common test scenarios
type TestScenario struct {
	Name           string
	Description    string
	Method         string
	Path           string
	Body           string
	Headers        map[string]string
	ExpectedStatus int
	RequiresAuth   bool
	RequiresAdmin  bool
}

// CommonTestScenarios provides a set of common test scenarios
var CommonTestScenarios = []TestScenario{
	{
		Name:           "health_check",
		Description:    "Health check endpoint should always be accessible",
		Method:         "GET",
		Path:           "/health",
		ExpectedStatus: 200,
		RequiresAuth:   false,
		RequiresAdmin:  false,
	},
	{
		Name:           "api_status",
		Description:    "API status endpoint should be accessible",
		Method:         "GET",
		Path:           "/api/status",
		ExpectedStatus: 200,
		RequiresAuth:   false,
		RequiresAdmin:  false,
	},
	{
		Name:           "plant_info_public",
		Description:    "Plant information should be publicly accessible",
		Method:         "GET",
		Path:           "/api/plant/",
		ExpectedStatus: 200,
		RequiresAuth:   false,
		RequiresAdmin:  false,
	},
	{
		Name:           "water_plant_protected",
		Description:    "Watering plant should require authentication",
		Method:         "POST",
		Path:           "/api/plant/water",
		ExpectedStatus: 401,
		RequiresAuth:   true,
		RequiresAdmin:  false,
	},
	{
		Name:           "admin_config_protected",
		Description:    "Admin config should require admin privileges",
		Method:         "GET",
		Path:           "/admin/config",
		ExpectedStatus: 401,
		RequiresAuth:   true,
		RequiresAdmin:  true,
	},
}

// SecurityTestCases provides test cases for security testing
var SecurityTestCases = []TestScenario{
	{
		Name:           "sql_injection_attempt",
		Description:    "Should handle SQL injection attempts safely",
		Method:         "POST",
		Path:           "/admin/users",
		Body:           `{"email": "test'; DROP TABLE users; --@example.com"}`,
		Headers:        TestHTTPHeaders.JSON,
		ExpectedStatus: 401, // Should be unauthorized before even processing
		RequiresAuth:   true,
		RequiresAdmin:  true,
	},
	{
		Name:           "xss_attempt",
		Description:    "Should handle XSS attempts safely",
		Method:         "POST",
		Path:           "/admin/users",
		Body:           `{"email": "<script>alert('xss')</script>@example.com"}`,
		Headers:        TestHTTPHeaders.JSON,
		ExpectedStatus: 401,
		RequiresAuth:   true,
		RequiresAdmin:  true,
	},
	{
		Name:           "oversized_request",
		Description:    "Should handle oversized requests",
		Method:         "POST",
		Path:           "/admin/users",
		Body:           `{"email": "` + string(make([]byte, 10000)) + `@example.com"}`,
		Headers:        TestHTTPHeaders.JSON,
		ExpectedStatus: 401, // Auth will fail before processing large request
		RequiresAuth:   true,
		RequiresAdmin:  true,
	},
}
