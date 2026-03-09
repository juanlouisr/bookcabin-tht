package providers

import (
	"context"
	"os"
	"testing"
	"time"

	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/transport"
)

func TestLionAirProvider_Search_Success(t *testing.T) {
	// Load test data from spec file
	data, err := os.ReadFile("../../spec/lion_air_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockTransport := transport.NewMockTransportFromBytes(data, 10, 20, 1.0)
	provider := NewLionAirProvider(mockTransport)

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

	// Should have 3 flights
	if len(flights) != 3 {
		t.Errorf("Expected 3 flights, got %d", len(flights))
	}

	// Verify first flight details
	if len(flights) > 0 {
		flight := flights[0]
		if flight.Provider != "Lion Air" {
			t.Errorf("Expected provider 'Lion Air', got '%s'", flight.Provider)
		}
		if flight.Airline.Code != "JT" {
			t.Errorf("Expected airline code 'JT', got '%s'", flight.Airline.Code)
		}
	}
}

func TestLionAirProvider_Search_UnsuccessfulResponse(t *testing.T) {
	provider := NewLionAirProvider(nil)
	responseWithError := []byte(`{"success": false, "data": {"available_flights": []}}`)

	_, err := provider.parseResponse(responseWithError, "CGK", "DPS")
	if err == nil {
		t.Error("Expected error for unsuccessful response, got nil")
	}
}

func TestLionAirProvider_Search_DirectAndConnecting(t *testing.T) {
	// Load test data
	data, err := os.ReadFile("../../spec/lion_air_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockTransport := transport.NewMockTransportFromBytes(data, 10, 20, 1.0)
	provider := NewLionAirProvider(mockTransport)

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

	if directCount != 2 {
		t.Errorf("Expected 2 direct flights, got %d", directCount)
	}
	if connectingCount != 1 {
		t.Errorf("Expected 1 connecting flight, got %d", connectingCount)
	}
}

func TestLionAirProvider_Name(t *testing.T) {
	provider := NewLionAirProvider(nil)
	if provider.Name() != "Lion Air" {
		t.Errorf("Expected name 'Lion Air', got '%s'", provider.Name())
	}
}

func TestLionAirProvider_ParseResponse_InvalidJSON(t *testing.T) {
	provider := NewLionAirProvider(nil)
	invalidJSON := []byte(`{invalid json}`)

	_, err := provider.parseResponse(invalidJSON, "CGK", "DPS")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLionAirProvider_ParseResponse_InvalidTimeFormat(t *testing.T) {
	provider := NewLionAirProvider(nil)
	responseWithInvalidTime := []byte(`{
		"success": true,
		"data": {
			"available_flights": [{
				"id": "JT740",
				"carrier": {
					"name": "Lion Air",
					"iata": "JT"
				},
				"route": {
					"from": {
						"code": "CGK",
						"name": "Soekarno-Hatta International",
						"city": "Jakarta"
					},
					"to": {
						"code": "DPS",
						"name": "Ngurah Rai International",
						"city": "Denpasar"
					}
				},
				"schedule": {
					"departure": "invalid-time",
					"departure_timezone": "Asia/Jakarta",
					"arrival": "2025-12-15T08:15:00",
					"arrival_timezone": "Asia/Makassar"
				},
				"flight_time": 105,
				"is_direct": true,
				"pricing": {
					"total": 950000,
					"currency": "IDR",
					"fare_type": "ECONOMY"
				},
				"seats_left": 45,
				"plane_type": "Boeing 737-900ER",
				"services": {
					"wifi_available": false,
					"meals_included": false,
					"baggage_allowance": {
						"cabin": "7 kg",
						"hold": "20 kg"
					}
				}
			}]
		}
	}`)

	flights, err := provider.parseResponse(responseWithInvalidTime, "CGK", "DPS")
	if err != nil {
		t.Fatalf("Expected no error (flight should be skipped), got: %v", err)
	}

	if len(flights) != 0 {
		t.Errorf("Expected 0 flights (invalid time should be skipped), got %d", len(flights))
	}
}

func TestLionAirProvider_Search_TransportError(t *testing.T) {
	mockTransport := transport.NewMockTransportFromBytes([]byte("{}"), 10, 20, 0.0)
	provider := NewLionAirProvider(mockTransport)

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

func TestLionAirProvider_Search_Timeout(t *testing.T) {
	data, err := os.ReadFile("../../spec/lion_air_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockTransport := transport.NewMockTransportFromBytes(data, 500, 600, 1.0)
	provider := NewLionAirProvider(mockTransport)

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

func TestLionAirProvider_Amenities(t *testing.T) {
	provider := NewLionAirProvider(nil)
	responseWithAmenities := []byte(`{
		"success": true,
		"data": {
			"available_flights": [
				{
					"id": "JT100",
					"carrier": {"name": "Lion Air", "iata": "JT"},
					"route": {
						"from": {"code": "CGK", "name": "CGK", "city": "Jakarta"},
						"to": {"code": "DPS", "name": "DPS", "city": "Denpasar"}
					},
					"schedule": {
						"departure": "2025-12-15T10:00:00",
						"departure_timezone": "Asia/Jakarta",
						"arrival": "2025-12-15T13:00:00",
						"arrival_timezone": "Asia/Makassar"
					},
					"flight_time": 120,
					"is_direct": true,
					"pricing": {"total": 1000000, "currency": "IDR", "fare_type": "ECONOMY"},
					"seats_left": 50,
					"plane_type": "Boeing 737",
					"services": {
						"wifi_available": true,
						"meals_included": true,
						"baggage_allowance": {"cabin": "7 kg", "hold": "20 kg"}
					}
				}
			]
		}
	}`)

	flights, err := provider.parseResponse(responseWithAmenities, "CGK", "DPS")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(flights) != 1 {
		t.Fatalf("Expected 1 flight, got %d", len(flights))
	}

	flight := flights[0]
	if len(flight.Amenities) != 2 {
		t.Errorf("Expected 2 amenities, got %d", len(flight.Amenities))
	}

	// Check for wifi and meal
	hasWifi := false
	hasMeal := false
	for _, a := range flight.Amenities {
		if a == "wifi" {
			hasWifi = true
		}
		if a == "meal" {
			hasMeal = true
		}
	}

	if !hasWifi {
		t.Error("Expected wifi amenity")
	}
	if !hasMeal {
		t.Error("Expected meal amenity")
	}
}

func TestParseLionAirResponse(t *testing.T) {
	validJSON := []byte(`{
		"success": true,
		"data": {
			"available_flights": [{"id": "JT740"}]
		}
	}`)

	response, err := ParseLionAirResponse(validJSON)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if len(response.Data.AvailableFlights) != 1 {
		t.Errorf("Expected 1 flight, got %d", len(response.Data.AvailableFlights))
	}

	// Test invalid JSON
	_, err = ParseLionAirResponse([]byte(`{invalid}`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
