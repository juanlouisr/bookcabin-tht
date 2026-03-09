package providers

import (
	"context"
	"os"
	"testing"
	"time"

	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/transport"
)

func TestBatikAirProvider_Search_Success(t *testing.T) {
	// Load test data from spec file
	data, err := os.ReadFile("../../spec/batik_air_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockTransport := transport.NewMockTransportFromBytes(data, 10, 20, 1.0)
	provider := NewBatikAirProvider(mockTransport)

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
		if flight.Provider != "Batik Air" {
			t.Errorf("Expected provider 'Batik Air', got '%s'", flight.Provider)
		}
		if flight.Airline.Code != "ID" {
			t.Errorf("Expected airline code 'ID', got '%s'", flight.Airline.Code)
		}
		// Verify price includes taxes
		if flight.Price.Amount != 1100000 {
			t.Errorf("Expected price 1100000 (including taxes), got %d", flight.Price.Amount)
		}
	}
}

func TestBatikAirProvider_Search_Non200Code(t *testing.T) {
	provider := NewBatikAirProvider(nil)
	responseWithError := []byte(`{"code": 500, "message": "Internal Server Error", "results": []}`)

	_, err := provider.parseResponse(responseWithError, "CGK", "DPS")
	if err == nil {
		t.Error("Expected error for non-200 code, got nil")
	}
}

func TestBatikAirProvider_Search_WithConnections(t *testing.T) {
	// Load test data
	data, err := os.ReadFile("../../spec/batik_air_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockTransport := transport.NewMockTransportFromBytes(data, 10, 20, 1.0)
	provider := NewBatikAirProvider(mockTransport)

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

	// Find the flight with connections
	var connectingFlight *models.UnifiedFlight
	for _, f := range flights {
		if f.Stops > 0 {
			connectingFlight = &f
			break
		}
	}

	if connectingFlight == nil {
		t.Error("Expected to find a connecting flight")
	} else {
		if connectingFlight.Stops != 1 {
			t.Errorf("Expected 1 stop, got %d", connectingFlight.Stops)
		}
	}
}

func TestBatikAirProvider_Name(t *testing.T) {
	provider := NewBatikAirProvider(nil)
	if provider.Name() != "Batik Air" {
		t.Errorf("Expected name 'Batik Air', got '%s'", provider.Name())
	}
}

func TestBatikAirProvider_ParseResponse_InvalidJSON(t *testing.T) {
	provider := NewBatikAirProvider(nil)
	invalidJSON := []byte(`{invalid json}`)

	_, err := provider.parseResponse(invalidJSON, "CGK", "DPS")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestBatikAirProvider_ParseResponse_InvalidTimeFormat(t *testing.T) {
	provider := NewBatikAirProvider(nil)
	responseWithInvalidTime := []byte(`{
		"code": 200,
		"message": "OK",
		"results": [{
			"flightNumber": "ID6514",
			"airlineName": "Batik Air",
			"airlineIATA": "ID",
			"origin": "CGK",
			"destination": "DPS",
			"departureDateTime": "invalid-time",
			"arrivalDateTime": "2025-12-15T10:00:00+0800",
			"travelTime": "1h 45m",
			"numberOfStops": 0,
			"fare": {
				"basePrice": 980000,
				"taxes": 120000,
				"totalPrice": 1100000,
				"currencyCode": "IDR",
				"class": "Y"
			},
			"seatsAvailable": 32,
			"aircraftModel": "Airbus A320",
			"baggageInfo": "7kg cabin, 20kg checked",
			"onboardServices": ["Snack", "Beverage"]
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

func TestBatikAirProvider_Search_TransportError(t *testing.T) {
	mockTransport := transport.NewMockTransportFromBytes([]byte("{}"), 10, 20, 0.0)
	provider := NewBatikAirProvider(mockTransport)

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

func TestBatikAirProvider_Search_Timeout(t *testing.T) {
	data, err := os.ReadFile("../../spec/batik_air_search_response.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	mockTransport := transport.NewMockTransportFromBytes(data, 500, 600, 1.0)
	provider := NewBatikAirProvider(mockTransport)

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

func TestBatikAirProvider_Amenities(t *testing.T) {
	provider := NewBatikAirProvider(nil)
	responseWithServices := []byte(`{
		"code": 200,
		"message": "OK",
		"results": [
			{
				"flightNumber": "ID6514",
				"airlineName": "Batik Air",
				"airlineIATA": "ID",
				"origin": "CGK",
				"destination": "DPS",
				"departureDateTime": "2025-12-15T07:15:00+0700",
				"arrivalDateTime": "2025-12-15T10:00:00+0800",
				"travelTime": "1h 45m",
				"numberOfStops": 0,
				"fare": {
					"basePrice": 980000,
					"taxes": 120000,
					"totalPrice": 1100000,
					"currencyCode": "IDR",
					"class": "Y"
				},
				"seatsAvailable": 32,
				"aircraftModel": "Airbus A320",
				"baggageInfo": "7kg cabin, 20kg checked",
				"onboardServices": ["Snack", "Beverage", "Entertainment", "Meal"]
			}
		]
	}`)

	flights, err := provider.parseResponse(responseWithServices, "CGK", "DPS")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(flights) != 1 {
		t.Fatalf("Expected 1 flight, got %d", len(flights))
	}

	flight := flights[0]
	if len(flight.Amenities) != 4 {
		t.Errorf("Expected 4 amenities, got %d", len(flight.Amenities))
	}
}

func TestParseBatikAirResponse(t *testing.T) {
	validJSON := []byte(`{
		"code": 200,
		"message": "OK",
		"results": [{"flightNumber": "ID6514"}]
	}`)

	response, err := ParseBatikAirResponse(validJSON)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.Code != 200 {
		t.Errorf("Expected code 200, got %d", response.Code)
	}

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}

	// Test invalid JSON
	_, err = ParseBatikAirResponse([]byte(`{invalid}`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
