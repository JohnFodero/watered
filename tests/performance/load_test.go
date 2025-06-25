package performance

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"watered/internal/auth"
	"watered/internal/handlers"
	"watered/internal/services"
	"watered/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
)

// LoadTestConfig holds configuration for load testing
type LoadTestConfig struct {
	Concurrency int           // Number of concurrent users
	Duration    time.Duration // Test duration
	RampUp      time.Duration // Time to ramp up to full concurrency
}

// LoadTestResults holds the results of a load test
type LoadTestResults struct {
	TotalRequests     int64
	SuccessfulReqs    int64
	FailedReqs        int64
	AvgResponseTime   time.Duration
	MaxResponseTime   time.Duration
	MinResponseTime   time.Duration
	RequestsPerSecond float64
}

// CreateLoadTestServer creates a server optimized for load testing
func CreateLoadTestServer() *httptest.Server {
	// Initialize storage
	store := storage.NewMemoryStorage()
	
	// Initialize services
	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)
	
	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService)
	plantHandlers := handlers.NewPlantHandlers(plantService, authService)

	// Create router
	r := chi.NewRouter()

	// Add basic middleware for performance testing
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Minimal middleware for performance testing
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/status", handlers.GetStatus)
		r.Get("/plant/", plantHandlers.GetPlantHandler)
		r.Get("/plant/status", plantHandlers.GetPlantStatusHandler)
		r.Get("/plant/timer", plantHandlers.GetPlantTimerHandler)
	})

	// Auth routes
	r.Get("/auth/status", authHandlers.StatusHandler)

	return httptest.NewServer(r)
}

// RunLoadTest executes a load test against the given endpoint
func RunLoadTest(t *testing.T, server *httptest.Server, endpoint string, config LoadTestConfig) *LoadTestResults {
	var (
		totalRequests   int64
		successfulReqs  int64
		failedReqs      int64
		totalTime       int64
		maxTime         int64
		minTime         int64 = int64(time.Hour) // Initialize to a large value
	)

	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	// Calculate ramp-up rate
	rampUpRate := time.Duration(int64(config.RampUp) / int64(config.Concurrency))
	
	var wg sync.WaitGroup
	startTime := time.Now()

	// Start workers with ramp-up
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		
		go func(workerID int) {
			defer wg.Done()
			
			// Ramp-up delay
			time.Sleep(time.Duration(workerID) * rampUpRate)
			
			// Create HTTP client for this worker
			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			
			for {
				select {
				case <-ctx.Done():
					return
				default:
					reqStart := time.Now()
					resp, err := client.Get(server.URL + endpoint)
					reqDuration := time.Since(reqStart)
					
					atomic.AddInt64(&totalRequests, 1)
					atomic.AddInt64(&totalTime, int64(reqDuration))
					
					// Update min/max response times
					for {
						current := atomic.LoadInt64(&maxTime)
						if int64(reqDuration) <= current {
							break
						}
						if atomic.CompareAndSwapInt64(&maxTime, current, int64(reqDuration)) {
							break
						}
					}
					
					for {
						current := atomic.LoadInt64(&minTime)
						if int64(reqDuration) >= current {
							break
						}
						if atomic.CompareAndSwapInt64(&minTime, current, int64(reqDuration)) {
							break
						}
					}
					
					if err != nil || resp.StatusCode != http.StatusOK {
						atomic.AddInt64(&failedReqs, 1)
						if err != nil {
							t.Logf("Worker %d: Request failed: %v", workerID, err)
						} else {
							t.Logf("Worker %d: Request failed with status: %d", workerID, resp.StatusCode)
						}
					} else {
						atomic.AddInt64(&successfulReqs, 1)
					}
					
					if resp != nil && resp.Body != nil {
						resp.Body.Close()
					}
					
					// Small delay to avoid overwhelming the server
					time.Sleep(1 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	results := &LoadTestResults{
		TotalRequests:     atomic.LoadInt64(&totalRequests),
		SuccessfulReqs:    atomic.LoadInt64(&successfulReqs),
		FailedReqs:        atomic.LoadInt64(&failedReqs),
		MaxResponseTime:   time.Duration(atomic.LoadInt64(&maxTime)),
		MinResponseTime:   time.Duration(atomic.LoadInt64(&minTime)),
		RequestsPerSecond: float64(atomic.LoadInt64(&totalRequests)) / totalDuration.Seconds(),
	}

	if results.TotalRequests > 0 {
		results.AvgResponseTime = time.Duration(atomic.LoadInt64(&totalTime) / results.TotalRequests)
	}

	return results
}

func TestHealthEndpointPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	server := CreateLoadTestServer()
	defer server.Close()

	config := LoadTestConfig{
		Concurrency: 10,
		Duration:    10 * time.Second,
		RampUp:      2 * time.Second,
	}

	t.Logf("Starting load test: %d concurrent users for %v", config.Concurrency, config.Duration)
	
	results := RunLoadTest(t, server, "/health", config)

	t.Logf("Load Test Results:")
	t.Logf("  Total Requests: %d", results.TotalRequests)
	t.Logf("  Successful: %d", results.SuccessfulReqs)
	t.Logf("  Failed: %d", results.FailedReqs)
	t.Logf("  Success Rate: %.2f%%", float64(results.SuccessfulReqs)/float64(results.TotalRequests)*100)
	t.Logf("  Requests/Second: %.2f", results.RequestsPerSecond)
	t.Logf("  Avg Response Time: %v", results.AvgResponseTime)
	t.Logf("  Min Response Time: %v", results.MinResponseTime)
	t.Logf("  Max Response Time: %v", results.MaxResponseTime)

	// Performance assertions
	require.Greater(t, results.TotalRequests, int64(50), "Should have processed at least 50 requests")
	require.Greater(t, float64(results.SuccessfulReqs)/float64(results.TotalRequests), 0.95, "Success rate should be > 95%")
	require.Less(t, results.AvgResponseTime, 100*time.Millisecond, "Average response time should be < 100ms")
	require.Greater(t, results.RequestsPerSecond, 50.0, "Should handle > 50 requests per second")
}

func TestPlantAPIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	server := CreateLoadTestServer()
	defer server.Close()

	endpoints := []string{
		"/api/plant/",
		"/api/plant/status",
		"/api/plant/timer",
		"/auth/status",
	}

	config := LoadTestConfig{
		Concurrency: 5,
		Duration:    5 * time.Second,
		RampUp:      1 * time.Second,
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			t.Logf("Testing endpoint: %s", endpoint)
			
			results := RunLoadTest(t, server, endpoint, config)

			t.Logf("Results for %s:", endpoint)
			t.Logf("  Requests/Second: %.2f", results.RequestsPerSecond)
			t.Logf("  Avg Response Time: %v", results.AvgResponseTime)
			t.Logf("  Success Rate: %.2f%%", float64(results.SuccessfulReqs)/float64(results.TotalRequests)*100)

			// Basic performance requirements
			require.Greater(t, float64(results.SuccessfulReqs)/float64(results.TotalRequests), 0.9, 
				"Success rate should be > 90% for %s", endpoint)
			require.Less(t, results.AvgResponseTime, 200*time.Millisecond, 
				"Average response time should be < 200ms for %s", endpoint)
		})
	}
}

func TestConcurrentUserScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	server := CreateLoadTestServer()
	defer server.Close()

	// Simulate realistic user behavior
	config := LoadTestConfig{
		Concurrency: 20,
		Duration:    15 * time.Second,
		RampUp:      3 * time.Second,
	}

	// Test mixed workload
	endpoints := []string{
		"/health",
		"/api/plant/",
		"/api/plant/status",
		"/auth/status",
	}

	var wg sync.WaitGroup
	results := make(map[string]*LoadTestResults)
	mu := sync.Mutex{}

	for _, endpoint := range endpoints {
		wg.Add(1)
		go func(ep string) {
			defer wg.Done()
			
			// Adjust concurrency per endpoint
			epConfig := config
			epConfig.Concurrency = config.Concurrency / len(endpoints)
			
			result := RunLoadTest(t, server, ep, epConfig)
			
			mu.Lock()
			results[ep] = result
			mu.Unlock()
		}(endpoint)
	}

	wg.Wait()

	// Analyze overall performance
	var totalReqs, totalSuccessful int64
	var totalRPS float64

	t.Logf("Mixed Workload Results:")
	for endpoint, result := range results {
		totalReqs += result.TotalRequests
		totalSuccessful += result.SuccessfulReqs
		totalRPS += result.RequestsPerSecond
		
		t.Logf("  %s: %.2f RPS, %.2f%% success", 
			endpoint, 
			result.RequestsPerSecond,
			float64(result.SuccessfulReqs)/float64(result.TotalRequests)*100)
	}

	overallSuccessRate := float64(totalSuccessful) / float64(totalReqs) * 100
	t.Logf("Overall: %.2f total RPS, %.2f%% success rate", totalRPS, overallSuccessRate)

	// Overall performance requirements
	require.Greater(t, overallSuccessRate, 90.0, "Overall success rate should be > 90%")
	require.Greater(t, totalRPS, 40.0, "Total RPS should be > 40")
}

func BenchmarkHealthEndpoint(b *testing.B) {
	server := CreateLoadTestServer()
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(server.URL + "/health")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				b.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
			resp.Body.Close()
		}
	})
}

func BenchmarkPlantAPI(b *testing.B) {
	server := CreateLoadTestServer()
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(server.URL + "/api/plant/")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				b.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
			resp.Body.Close()
		}
	})
}