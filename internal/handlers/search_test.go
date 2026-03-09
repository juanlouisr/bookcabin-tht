package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bookcabin/flight/internal/aggregator"
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
	time.Sleep(p.delay)
	return p.flights, nil
}

func createTestAggregator() *aggregator.Service {
	provider := &MockProvider{
		name: "TestProvider",
		flights: []models.UnifiedFlight{
			{
				ID:       "TEST001",
				Provider: "TestProvider",
				Airline:  models.AirlineInfo{Name: "Test Air", Code: "TST"},
				Departure: models.DepartureInfo{
					Airport:   "CGK",
					City:      "Jakarta",
					Datetime:  "2025-12-15T10:00:00+07:00",
					Timestamp: 1734249600,
				},
				Arrival: models.ArrivalInfo{
					Airport:   "DPS",
					City:      "Denpasar",
					Datetime:  "2025-12-15T13:00:00+08:00",
					Timestamp: 1734262800,
				},
				Duration:       models.DurationInfo{TotalMinutes: 120, Formatted: "2h 0m"},
				Stops:          0,
				Price:          models.PriceInfo{Amount: 1000000, Currency: "IDR"},
				AvailableSeats: 50,
				CabinClass:     "economy",
			},
		},
		delay: 0,
		fail:  false,
	}

	return aggregator.NewService([]models.Provider{provider}, cache.NewMemoryCache(), 5*time.Second)
}

func TestSearchHandler_HandleSearch_Success(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	requestBody := SearchHTTPRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleSearch(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response models.SearchResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Flights) != 1 {
		t.Errorf("Expected 1 flight, got %d", len(response.Flights))
	}
}

func TestSearchHandler_HandleSearch_MissingRequiredFields(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	testCases := []struct {
		name    string
		request SearchHTTPRequest
	}{
		{
			name:    "Missing origin",
			request: SearchHTTPRequest{Destination: "DPS", DepartureDate: "2025-12-15"},
		},
		{
			name:    "Missing destination",
			request: SearchHTTPRequest{Origin: "CGK", DepartureDate: "2025-12-15"},
		},
		{
			name:    "Missing departure date",
			request: SearchHTTPRequest{Origin: "CGK", Destination: "DPS"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.request)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.HandleSearch(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
			}
		})
	}
}

func TestSearchHandler_HandleSearch_InvalidJSON(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewBuffer([]byte(`{invalid json}`)))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleSearch(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestSearchHandler_HandleSearch_WrongMethod(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	rr := httptest.NewRecorder()
	handler.HandleSearch(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestSearchHandler_HandleSearch_WithFilters(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	maxPrice := 1500000
	requestBody := SearchHTTPRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Filters: FilterOptions{
			MaxPrice: &maxPrice,
		},
		SortBy: "price_asc",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleSearch(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestSearchHandler_HandleHealth(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.HandleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
}

func TestSearchHandler_HandleMultiCitySearch_Success(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	requestBody := MultiCitySearchHTTPRequest{
		Legs: []models.MultiCityLeg{
			{Origin: "CGK", Destination: "DPS", DepartureDate: "2025-12-15"},
			{Origin: "DPS", Destination: "CGK", DepartureDate: "2025-12-20"},
		},
		Passengers: 1,
		CabinClass: "economy",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/multi-city", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleMultiCitySearch(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	legs, ok := response["legs"].([]interface{})
	if !ok {
		t.Fatal("Expected 'legs' to be an array")
	}

	if len(legs) != 2 {
		t.Errorf("Expected 2 legs, got %d", len(legs))
	}
}

func TestSearchHandler_HandleMultiCitySearch_NoLegs(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	requestBody := MultiCitySearchHTTPRequest{
		Legs:       []models.MultiCityLeg{},
		Passengers: 1,
		CabinClass: "economy",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/multi-city", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleMultiCitySearch(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestSearchHandler_HandleMultiCitySearch_InvalidLeg(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	requestBody := MultiCitySearchHTTPRequest{
		Legs: []models.MultiCityLeg{
			{Origin: "", Destination: "DPS", DepartureDate: "2025-12-15"}, // Missing origin
		},
		Passengers: 1,
		CabinClass: "economy",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search/multi-city", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleMultiCitySearch(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestConvertSortOption(t *testing.T) {
	tests := []struct {
		input    string
		expected models.SortOption
	}{
		{"price_asc", models.SortByPriceAsc},
		{"price", models.SortByPriceAsc},
		{"cheapest", models.SortByPriceAsc},
		{"price_desc", models.SortByPriceDesc},
		{"duration_asc", models.SortByDurationAsc},
		{"duration", models.SortByDurationAsc},
		{"fastest", models.SortByDurationAsc},
		{"duration_desc", models.SortByDurationDesc},
		{"departure_asc", models.SortByDepartureAsc},
		{"departure", models.SortByDepartureAsc},
		{"departure_desc", models.SortByDepartureDesc},
		{"arrival_asc", models.SortByArrivalAsc},
		{"arrival", models.SortByArrivalAsc},
		{"arrival_desc", models.SortByArrivalDesc},
		{"best_value", models.SortByBestValue},
		{"bestvalue", models.SortByBestValue},
		{"recommended", models.SortByBestValue},
		{"unknown", models.SortByPriceAsc}, // Default
		{"", models.SortByPriceAsc},        // Default
	}

	for _, test := range tests {
		result := convertSortOption(test.input)
		if result != test.expected {
			t.Errorf("convertSortOption(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestConvertFilterOptions(t *testing.T) {
	maxPrice := 1000000
	minPrice := 500000
	maxStops := 1
	departureAfter := "08:00"
	departureBefore := "18:00"
	arrivalAfter := "10:00"
	arrivalBefore := "22:00"
	maxDuration := 300

	httpOpts := FilterOptions{
		MaxPrice:        &maxPrice,
		MinPrice:        &minPrice,
		MaxStops:        &maxStops,
		Airlines:        []string{"Garuda Indonesia"},
		DepartureAfter:  &departureAfter,
		DepartureBefore: &departureBefore,
		ArrivalAfter:    &arrivalAfter,
		ArrivalBefore:   &arrivalBefore,
		MaxDurationMins: &maxDuration,
	}

	result := convertFilterOptions(httpOpts)

	if result.MaxPrice == nil || *result.MaxPrice != maxPrice {
		t.Error("MaxPrice not converted correctly")
	}
	if result.MinPrice == nil || *result.MinPrice != minPrice {
		t.Error("MinPrice not converted correctly")
	}
	if result.MaxStops == nil || *result.MaxStops != maxStops {
		t.Error("MaxStops not converted correctly")
	}
	if len(result.Airlines) != 1 || result.Airlines[0] != "Garuda Indonesia" {
		t.Error("Airlines not converted correctly")
	}
	if result.DepartureAfter == nil {
		t.Error("DepartureAfter not converted")
	}
	if result.DepartureBefore == nil {
		t.Error("DepartureBefore not converted")
	}
	if result.ArrivalAfter == nil {
		t.Error("ArrivalAfter not converted")
	}
	if result.ArrivalBefore == nil {
		t.Error("ArrivalBefore not converted")
	}
	if result.MaxDurationMins == nil || *result.MaxDurationMins != maxDuration {
		t.Error("MaxDurationMins not converted correctly")
	}
}

func TestSearchHandler_DefaultValues(t *testing.T) {
	handler := NewSearchHandler(createTestAggregator())

	// Request with missing passengers and cabin_class
	requestBody := map[string]string{
		"origin":         "CGK",
		"destination":    "DPS",
		"departure_date": "2025-12-15",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleSearch(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}
