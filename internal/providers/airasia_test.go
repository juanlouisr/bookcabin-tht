package providers

import (
	"context"
	"os"
	"testing"
	"time"

	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/transport"
)

func TestAirAsiaProvider_Search_Success(t *testing.T) {
	// Load test data from spec file
	data, err := os.ReadFile("../../spec/airasia_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockTransport := transport.NewMockTransportFromBytes(data, 10, 20, 1.0)
	provider := NewAirAsiaProvider(mockTransport)

	request := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have 4 flights
	if len(flights) != 4 {
		t.Errorf("Expected 4 flights, got %d", len(flights))
	}

	// Verify first flight details
	if len(flights) > 0 {
		flight := flights[0]
		if flight.Provider != "AirAsia" {
			t.Errorf("Expected provider 'AirAsia', got '%s'", flight.Provider)
		}
		if flight.Airline.Code != "QZ" {
			t.Errorf("Expected airline code 'QZ', got '%s'", flight.Airline.Code)
		}
		if flight.CabinClass != "economy" {
			t.Errorf("Expected cabin class 'economy', got '%s'", flight.CabinClass)
		}
	}
}

func TestAirAsiaProvider_Search_NonOkStatus(t *testing.T) {
	provider := NewAirAsiaProvider(nil)
	responseWithError := []byte(`{"status": "error", "flights": []}`)

	_, err := provider.parseResponse(responseWithError, "CGK", "DPS")
	if err == nil {
		t.Error("Expected error for non-ok status, got nil")
	}
}

func TestAirAsiaProvider_Search_DirectAndConnecting(t *testing.T) {
	// Load test data
	data, err := os.ReadFile("../../spec/airasia_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockTransport := transport.NewMockTransportFromBytes(data, 10, 20, 1.0)
	provider := NewAirAsiaProvider(mockTransport)

	request := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, request)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Count direct and connecting flights
	directCount := 0
	connectingCount := 0
	for _, f := range flights {
		if f.Stops == 0 {
			directCount++
		} else {
			connectingCount++
		}
	}

	if directCount != 3 {
		t.Errorf("Expected 3 direct flights, got %d", directCount)
	}
	if connectingCount != 1 {
		t.Errorf("Expected 1 connecting flight, got %d", connectingCount)
	}
}

func TestAirAsiaProvider_Name(t *testing.T) {
	provider := NewAirAsiaProvider(nil)
	if provider.Name() != "AirAsia" {
		t.Errorf("Expected name 'AirAsia', got '%s'", provider.Name())
	}
}

func TestAirAsiaProvider_ParseResponse_InvalidJSON(t *testing.T) {
	provider := NewAirAsiaProvider(nil)
	invalidJSON := []byte(`{invalid json}`)

	_, err := provider.parseResponse(invalidJSON, "CGK", "DPS")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestAirAsiaProvider_ParseResponse_InvalidTimeFormat(t *testing.T) {
	provider := NewAirAsiaProvider(nil)
	responseWithInvalidTime := []byte(`{
		"status": "ok",
		"flights": [{
			"flight_code": "QZ520",
			"airline": "AirAsia",
			"from_airport": "CGK",
			"to_airport": "DPS",
			"depart_time": "invalid-time",
			"arrive_time": "2025-12-15T07:25:00+08:00",
			"duration_hours": 1.67,
			"direct_flight": true,
			"price_idr": 650000,
			"seats": 67,
			"cabin_class": "economy",
			"baggage_note": "Cabin baggage only"
		}]
	}`)

	flights, err := provider.parseResponse(responseWithInvalidTime, "CGK", "DPS")
	if err != nil {
		t.Fatalf("Expected no error (flight should be skipped), got: %v", err)
	}

	if len(flights) != 0 {
		t.Errorf("Expected 0 flights (invalid time should be skipped), got %d", len(flights))
	}
}

func TestAirAsiaProvider_Search_TransportError(t *testing.T) {
	mockTransport := transport.NewMockTransportFromBytes([]byte("{}"), 10, 20, 0.0)
	provider := NewAirAsiaProvider(mockTransport)

	request := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	ctx := context.Background()
	_, err := provider.Search(ctx, request)

	if err == nil {
		t.Error("Expected error for failed transport, got nil")
	}
}

func TestAirAsiaProvider_Search_Timeout(t *testing.T) {
	data, err := os.ReadFile("../../spec/airasia_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	// Create slow mock transport
	mockTransport := transport.NewMockTransportFromBytes(data, 500, 600, 1.0)
	provider := NewAirAsiaProvider(mockTransport)

	request := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = provider.Search(ctx, request)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestAirAsiaProvider_ParseResponse_WithStops(t *testing.T) {
	provider := NewAirAsiaProvider(nil)
	responseWithStops := []byte(`{
		"status": "ok",
		"flights": [{
			"flight_code": "QZ7250",
			"airline": "AirAsia",
			"from_airport": "CGK",
			"to_airport": "DPS",
			"depart_time": "2025-12-15T15:15:00+07:00",
			"arrive_time": "2025-12-15T20:35:00+08:00",
			"duration_hours": 4.33,
			"direct_flight": false,
			"stops": [
				{
					"airport": "SOC",
					"wait_time_minutes": 95
				}
			],
			"price_idr": 485000,
			"seats": 88,
			"cabin_class": "economy",
			"baggage_note": "Cabin baggage only"
		}]
	}`)

	flights, err := provider.parseResponse(responseWithStops, "CGK", "DPS")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(flights) != 1 {
		t.Fatalf("Expected 1 flight, got %d", len(flights))
	}

	flight := flights[0]
	if flight.Stops != 1 {
		t.Errorf("Expected 1 stop, got %d", flight.Stops)
	}

	// 4.33 hours = 259.8 minutes, truncated to 259
	expectedDuration := 259
	if flight.Duration.TotalMinutes != expectedDuration {
		t.Errorf("Expected duration %d minutes, got %d", expectedDuration, flight.Duration.TotalMinutes)
	}
}

func TestParseAirAsiaResponse(t *testing.T) {
	validJSON := []byte(`{
		"status": "ok",
		"flights": [
			{
				"flight_code": "QZ520",
				"airline": "AirAsia",
				"from_airport": "CGK",
				"to_airport": "DPS",
				"depart_time": "2025-12-15T04:45:00+07:00",
				"arrive_time": "2025-12-15T07:25:00+08:00",
				"duration_hours": 1.67,
				"direct_flight": true,
				"price_idr": 650000,
				"seats": 67,
				"cabin_class": "economy",
				"baggage_note": "Cabin baggage only"
			}
		]
	}`)

	response, err := ParseAirAsiaResponse(validJSON)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	if len(response.Flights) != 1 {
		t.Errorf("Expected 1 flight, got %d", len(response.Flights))
	}

	// Test invalid JSON
	_, err = ParseAirAsiaResponse([]byte(`{invalid}`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
