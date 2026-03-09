package aggregator

import (
	"context"
	"testing"
	"time"

	"bookcabin/flight/internal/cache"
	"bookcabin/flight/internal/models"
)

// MockProvider implements models.Provider for testing
type MockProvider struct {
	name    string
	flights []models.UnifiedFlight
	delay   time.Duration
	fail    bool
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) Search(ctx context.Context, request models.SearchRequest) ([]models.UnifiedFlight, error) {
	if p.fail {
		return nil, context.DeadlineExceeded
	}

	select {
	case <-time.After(p.delay):
		return p.flights, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func createMockProvider(name string, flights []models.UnifiedFlight, delay time.Duration, fail bool) *MockProvider {
	return &MockProvider{
		name:    name,
		flights: flights,
		delay:   delay,
		fail:    fail,
	}
}

func TestService_Search_Success(t *testing.T) {
	// Create mock providers
	provider1 := createMockProvider("Provider1", []models.UnifiedFlight{
		{ID: "f1", Price: models.PriceInfo{Amount: 1000000}},
	}, 10*time.Millisecond, false)

	provider2 := createMockProvider("Provider2", []models.UnifiedFlight{
		{ID: "f2", Price: models.PriceInfo{Amount: 800000}},
	}, 10*time.Millisecond, false)

	providers := []models.Provider{provider1, provider2}
	svc := NewService(providers, nil, 5*time.Second)

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	ctx := context.Background()
	resp := svc.Search(ctx, req)

	if resp.Metadata.TotalResults != 2 {
		t.Errorf("Expected 2 total results, got %d", resp.Metadata.TotalResults)
	}
	if resp.Metadata.ProvidersQueried != 2 {
		t.Errorf("Expected 2 providers queried, got %d", resp.Metadata.ProvidersQueried)
	}
	if resp.Metadata.ProvidersSucceeded != 2 {
		t.Errorf("Expected 2 providers succeeded, got %d", resp.Metadata.ProvidersSucceeded)
	}
	if resp.Metadata.ProvidersFailed != 0 {
		t.Errorf("Expected 0 providers failed, got %d", resp.Metadata.ProvidersFailed)
	}
}

func TestService_Search_WithFailure(t *testing.T) {
	// Create mock providers - one fails
	provider1 := createMockProvider("Provider1", []models.UnifiedFlight{
		{ID: "f1", Price: models.PriceInfo{Amount: 1000000}},
	}, 10*time.Millisecond, false)

	provider2 := createMockProvider("Provider2", nil, 10*time.Millisecond, true) // Fails

	providers := []models.Provider{provider1, provider2}
	svc := NewService(providers, nil, 5*time.Second)

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	ctx := context.Background()
	resp := svc.Search(ctx, req)

	if resp.Metadata.TotalResults != 1 {
		t.Errorf("Expected 1 total result, got %d", resp.Metadata.TotalResults)
	}
	if resp.Metadata.ProvidersSucceeded != 1 {
		t.Errorf("Expected 1 provider succeeded, got %d", resp.Metadata.ProvidersSucceeded)
	}
	if resp.Metadata.ProvidersFailed != 1 {
		t.Errorf("Expected 1 provider failed, got %d", resp.Metadata.ProvidersFailed)
	}
}

func TestService_Search_WithCache(t *testing.T) {
	provider := createMockProvider("Provider1", []models.UnifiedFlight{
		{ID: "f1", Price: models.PriceInfo{Amount: 1000000}},
	}, 10*time.Millisecond, false)

	providers := []models.Provider{provider}
	flightCache := cache.NewMemoryCache()
	svc := NewService(providers, flightCache, 5*time.Second)

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		CabinClass:    "economy",
	}

	ctx := context.Background()

	// First search - cache miss
	resp1 := svc.Search(ctx, req)
	if resp1.Metadata.CacheHit {
		t.Error("Expected cache miss on first request")
	}

	// Second search - cache hit
	resp2 := svc.Search(ctx, req)
	if !resp2.Metadata.CacheHit {
		t.Error("Expected cache hit on second request")
	}
}

func TestService_Search_Timeout(t *testing.T) {
	// Create slow provider
	provider := createMockProvider("SlowProvider", []models.UnifiedFlight{
		{ID: "f1", Price: models.PriceInfo{Amount: 1000000}},
	}, 1*time.Second, false)

	providers := []models.Provider{provider}
	svc := NewService(providers, nil, 100*time.Millisecond) // Short timeout

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	ctx := context.Background()
	resp := svc.Search(ctx, req)

	if resp.Metadata.ProvidersFailed != 1 {
		t.Errorf("Expected 1 provider failed (timeout), got %d", resp.Metadata.ProvidersFailed)
	}
}

func TestCalculateBestValueScores(t *testing.T) {
	flights := []models.UnifiedFlight{
		{ID: "f1", Price: models.PriceInfo{Amount: 1000000}, Duration: models.DurationInfo{TotalMinutes: 120}, Stops: 0},
		{ID: "f2", Price: models.PriceInfo{Amount: 800000}, Duration: models.DurationInfo{TotalMinutes: 180}, Stops: 1},
		{ID: "f3", Price: models.PriceInfo{Amount: 1500000}, Duration: models.DurationInfo{TotalMinutes: 90}, Stops: 0},
	}

	result := calculateBestValueScores(flights)

	if len(result) != 3 {
		t.Fatalf("Expected 3 flights, got %d", len(result))
	}

	// All flights should have a best value score assigned
	for _, f := range result {
		if f.BestValueScore < 0 || f.BestValueScore > 1 {
			t.Errorf("Flight %s has invalid best value score: %f", f.ID, f.BestValueScore)
		}
	}
}

func TestGenerateCacheKey(t *testing.T) {
	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		CabinClass:    "economy",
	}

	key := generateCacheKey(req)
	expected := "CGK_DPS_2025-12-15_economy"

	if key != expected {
		t.Errorf("Expected cache key %s, got %s", expected, key)
	}
}
