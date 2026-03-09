package providers

import (
	"context"
	"os"
	"testing"
	"time"

	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/transport"
)

func TestGarudaProvider_Search_Success(t *testing.T) {
	// Load test data from spec file
	data, err := os.ReadFile("../../spec/garuda_indonesia_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	// Create mock transport
	mockTransport := transport.NewMockTransportFromBytes(data, 10, 20, 1.0)
	provider := NewGarudaProvider(mockTransport)

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

	// Should have at least 2 flights (GA400 and GA410, GA315 goes to SUB)
	if len(flights) < 2 {
		t.Errorf("Expected at least 2 flights, got %d", len(flights))
	}

	// Verify first flight details
	if len(flights) > 0 {
		flight := flights[0]
		if flight.Provider != "Garuda Indonesia" {
			t.Errorf("Expected provider 'Garuda Indonesia', got '%s'", flight.Provider)
		}
		if flight.Airline.Code != "GA" {
			t.Errorf("Expected airline code 'GA', got '%s'", flight.Airline.Code)
		}
		if flight.Departure.Airport != "CGK" {
			t.Errorf("Expected departure airport 'CGK', got '%s'", flight.Departure.Airport)
		}
		if flight.Arrival.Airport != "DPS" {
			t.Errorf("Expected arrival airport 'DPS', got '%s'", flight.Arrival.Airport)
		}
	}
}

func TestGarudaProvider_Search_NoMatchingFlights(t *testing.T) {
	// Load test data from spec file
	data, err := os.ReadFile("../../spec/garuda_indonesia_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockTransport := transport.NewMockTransportFromBytes(data, 10, 20, 1.0)
	provider := NewGarudaProvider(mockTransport)

	// Search for flights that don't exist in the test data
	request := models.SearchRequest{
		Origin:        "JFK",
		Destination:   "LAX",
		DepartureDate: "2025-12-15",
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(flights) != 0 {
		t.Errorf("Expected 0 flights for non-matching route, got %d", len(flights))
	}
}

func TestGarudaProvider_Search_TransportError(t *testing.T) {
	// Create mock transport with 0% success rate
	mockTransport := transport.NewMockTransportFromBytes([]byte("{}"), 10, 20, 0.0)
	provider := NewGarudaProvider(mockTransport)

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

func TestGarudaProvider_Name(t *testing.T) {
	provider := NewGarudaProvider(nil)
	if provider.Name() != "Garuda Indonesia" {
		t.Errorf("Expected name 'Garuda Indonesia', got '%s'", provider.Name())
	}
}

func TestGarudaProvider_ParseResponse_InvalidJSON(t *testing.T) {
	provider := NewGarudaProvider(nil)
	invalidJSON := []byte(`{invalid json}`)

	_, err := provider.parseResponse(invalidJSON, "CGK", "DPS")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGarudaProvider_ParseResponse_EmptyFlights(t *testing.T) {
	provider := NewGarudaProvider(nil)
	emptyResponse := []byte(`{"status": "success", "flights": []}`)

	flights, err := provider.parseResponse(emptyResponse, "CGK", "DPS")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(flights) != 0 {
		t.Errorf("Expected 0 flights, got %d", len(flights))
	}
}

func TestGarudaProvider_ParseResponse_WithSegments(t *testing.T) {
	provider := NewGarudaProvider(nil)
	responseWithSegments := []byte(`{
		"status": "success",
		"flights": [{
			"flight_id": "GA315",
			"airline": "Garuda Indonesia",
			"airline_code": "GA",
			"departure": {
				"airport": "CGK",
				"city": "Jakarta",
				"time": "2025-12-15T14:00:00+07:00",
				"terminal": "3"
			},
			"arrival": {
				"airport": "DPS",
				"city": "Denpasar",
				"time": "2025-12-15T18:45:00+08:00",
				"terminal": "I"
			},
			"duration_minutes": 165,
			"stops": 1,
			"aircraft": "Boeing 737",
			"price": {
				"amount": 1850000,
				"currency": "IDR"
			},
			"available_seats": 22,
			"fare_class": "economy",
			"baggage": {
				"carry_on": 1,
				"checked": 2
			},
			"segments": [
				{
					"flight_number": "GA315",
					"departure": {
						"airport": "CGK",
						"time": "2025-12-15T14:00:00+07:00"
					},
					"arrival": {
						"airport": "SUB",
						"time": "2025-12-15T15:30:00+07:00"
					},
					"duration_minutes": 90,
					"layover_minutes": 105
				},
				{
					"flight_number": "GA332",
					"departure": {
						"airport": "SUB",
						"time": "2025-12-15T17:15:00+07:00"
					},
					"arrival": {
						"airport": "DPS",
						"time": "2025-12-15T18:45:00+08:00"
					},
					"duration_minutes": 90
				}
			]
		}]
	}`)

	flights, err := provider.parseResponse(responseWithSegments, "CGK", "DPS")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(flights) != 1 {
		t.Fatalf("Expected 1 flight, got %d", len(flights))
	}

	flight := flights[0]
	if len(flight.Segments) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(flight.Segments))
	}

	if flight.Stops != 1 {
		t.Errorf("Expected 1 stop, got %d", flight.Stops)
	}
}

func TestGarudaProvider_ParseResponse_InvalidTimeFormat(t *testing.T) {
	provider := NewGarudaProvider(nil)
	responseWithInvalidTime := []byte(`{
		"status": "success",
		"flights": [{
			"flight_id": "GA400",
			"airline": "Garuda Indonesia",
			"airline_code": "GA",
			"departure": {
				"airport": "CGK",
				"city": "Jakarta",
				"time": "invalid-time-format",
				"terminal": "3"
			},
			"arrival": {
				"airport": "DPS",
				"city": "Denpasar",
				"time": "2025-12-15T08:50:00+08:00",
				"terminal": "I"
			},
			"duration_minutes": 110,
			"stops": 0,
			"aircraft": "Boeing 737-800",
			"price": {
				"amount": 1250000,
				"currency": "IDR"
			},
			"available_seats": 28,
			"fare_class": "economy",
			"baggage": {
				"carry_on": 1,
				"checked": 2
			}
		}]
	}`)

	flights, err := provider.parseResponse(responseWithInvalidTime, "CGK", "DPS")
	if err != nil {
		t.Fatalf("Expected no error (flight should be skipped), got: %v", err)
	}

	// Flight with invalid time should be skipped
	if len(flights) != 0 {
		t.Errorf("Expected 0 flights (invalid time should be skipped), got %d", len(flights))
	}
}

func TestGarudaProvider_Search_Timeout(t *testing.T) {
	// Load test data
	data, err := os.ReadFile("../../spec/garuda_indonesia_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	// Create slow mock transport (500ms delay)
	mockTransport := transport.NewMockTransportFromBytes(data, 500, 600, 1.0)
	provider := NewGarudaProvider(mockTransport)

	request := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = provider.Search(ctx, request)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestFormatBaggageCount(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "Not included"},
		{1, "1 piece"},
		{2, "2 pieces"},
		{3, "2 pieces"},
	}

	for _, test := range tests {
		result := formatBaggageCount(test.input)
		if result != test.expected {
			t.Errorf("formatBaggageCount(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}
}
