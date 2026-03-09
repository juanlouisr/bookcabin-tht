package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		minutes  int
		expected string
	}{
		{30, "30m"},
		{60, "1h"},
		{90, "1h 30m"},
		{120, "2h"},
		{150, "2h 30m"},
		{0, "0m"},
		{45, "45m"},
		{180, "3h"},
	}

	for _, test := range tests {
		result := FormatDuration(test.minutes)
		assert.Equal(t, test.expected, result, "FormatDuration(%d) should return %s", test.minutes, test.expected)
	}
}

func TestParseDurationString(t *testing.T) {
	tests := []struct {
		duration string
		expected int
	}{
		{"1h 30m", 90},
		{"2h 0m", 120},
		{"0h 45m", 45},
		{"3h 15m", 195},
		{"0h 30m", 30},
		{"", 0},        // Empty string
		{"invalid", 0}, // Invalid format
	}

	for _, test := range tests {
		result := ParseDurationString(test.duration)
		assert.Equal(t, test.expected, result, "ParseDurationString(%s) should return %d", test.duration, test.expected)
	}
}

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		amount   int
		expected string
	}{
		{500, "500"},
		{1000, "1.000"},
		{1500000, "1.500.000"},
		{1250000, "1.250.000"},
		{999, "999"},
		{1000000, "1.000.000"},
		{0, "0"},
		{999999999, "999.999.999"},
	}

	for _, test := range tests {
		result := FormatPrice(test.amount)
		assert.Equal(t, test.expected, result, "FormatPrice(%d) should return %s", test.amount, test.expected)
	}
}

func TestSearchRequest_Validation(t *testing.T) {
	departureDate := "2025-12-20"
	req := SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		ReturnDate:    &departureDate,
		Passengers:    2,
		CabinClass:    "business",
	}

	assert.Equal(t, "CGK", req.Origin)
	assert.Equal(t, "DPS", req.Destination)
	assert.Equal(t, "2025-12-15", req.DepartureDate)
	assert.NotNil(t, req.ReturnDate)
	assert.Equal(t, "2025-12-20", *req.ReturnDate)
	assert.Equal(t, 2, req.Passengers)
	assert.Equal(t, "business", req.CabinClass)
}

func TestUnifiedFlight_Structure(t *testing.T) {
	aircraft := "Boeing 737-800"
	flight := UnifiedFlight{
		ID:           "GA400_Garuda Indonesia",
		Provider:     "Garuda Indonesia",
		FlightNumber: "GA400",
		Airline: AirlineInfo{
			Name: "Garuda Indonesia",
			Code: "GA",
		},
		Departure: DepartureInfo{
			Airport:   "CGK",
			City:      "Jakarta",
			Datetime:  "2025-12-15T06:00:00+07:00",
			Timestamp: 1734249600,
			Terminal:  "3",
		},
		Arrival: ArrivalInfo{
			Airport:   "DPS",
			City:      "Denpasar",
			Datetime:  "2025-12-15T08:50:00+08:00",
			Timestamp: 1734258600,
			Terminal:  "I",
		},
		Duration: DurationInfo{
			TotalMinutes: 110,
			Formatted:    "1h 50m",
		},
		Stops:          0,
		Price:          PriceInfo{Amount: 1250000, Currency: "IDR"},
		AvailableSeats: 28,
		CabinClass:     "economy",
		Aircraft:       &aircraft,
		Amenities:      []string{"wifi", "meal", "entertainment"},
		Baggage: BaggageInfo{
			CarryOn: "1 piece",
			Checked: "2 pieces",
		},
		BestValueScore: 0.75,
	}

	assert.Equal(t, "GA400_Garuda Indonesia", flight.ID)
	assert.Equal(t, "GA", flight.Airline.Code)
	assert.Equal(t, "1h 50m", flight.Duration.Formatted)
	assert.Len(t, flight.Amenities, 3)
	assert.Contains(t, flight.Amenities, "wifi")
	assert.Contains(t, flight.Amenities, "meal")
	assert.Contains(t, flight.Amenities, "entertainment")
	assert.NotNil(t, flight.Aircraft)
	assert.Equal(t, "Boeing 737-800", *flight.Aircraft)
}

func TestMetadata_Structure(t *testing.T) {
	metadata := Metadata{
		TotalResults:       15,
		ProvidersQueried:   4,
		ProvidersSucceeded: 4,
		ProvidersFailed:    0,
		SearchTimeMs:       285,
		CacheHit:           false,
	}

	assert.Equal(t, 15, metadata.TotalResults)
	assert.Equal(t, 4, metadata.ProvidersQueried)
	assert.Equal(t, 4, metadata.ProvidersSucceeded)
	assert.Equal(t, 0, metadata.ProvidersFailed)
	assert.Equal(t, int64(285), metadata.SearchTimeMs)
	assert.False(t, metadata.CacheHit)
}

func TestSearchResponse_Structure(t *testing.T) {
	response := SearchResponse{
		SearchCriteria: SearchCriteria{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: "2025-12-15",
			Passengers:    1,
			CabinClass:    "economy",
		},
		Metadata: Metadata{
			TotalResults: 10,
		},
		Flights: []UnifiedFlight{
			{ID: "flight1"},
			{ID: "flight2"},
		},
	}

	assert.Equal(t, "CGK", response.SearchCriteria.Origin)
	assert.Equal(t, "DPS", response.SearchCriteria.Destination)
	assert.Equal(t, 10, response.Metadata.TotalResults)
	assert.Len(t, response.Flights, 2)
}

func TestFlightSegment_Structure(t *testing.T) {
	layoverMinutes := 90
	segment := FlightSegment{
		FlightNumber:    "GA315",
		DurationMinutes: 90,
		LayoverMinutes:  &layoverMinutes,
		Departure: SegmentPoint{
			Airport:   "CGK",
			Time:      "2025-12-15T14:00:00+07:00",
			Timestamp: 1734267600,
		},
		Arrival: SegmentPoint{
			Airport:   "SUB",
			Time:      "2025-12-15T15:30:00+07:00",
			Timestamp: 1734273000,
		},
	}

	assert.Equal(t, "GA315", segment.FlightNumber)
	assert.NotNil(t, segment.LayoverMinutes)
	assert.Equal(t, 90, *segment.LayoverMinutes)
	assert.Equal(t, "CGK", segment.Departure.Airport)
	assert.Equal(t, "SUB", segment.Arrival.Airport)
}

func TestFilterOptions_Structure(t *testing.T) {
	maxPrice := 1500000
	minPrice := 500000
	maxStops := 1

	opts := FilterOptions{
		MaxPrice: &maxPrice,
		MinPrice: &minPrice,
		MaxStops: &maxStops,
		Airlines: []string{"Garuda Indonesia", "Lion Air"},
	}

	assert.NotNil(t, opts.MaxPrice)
	assert.Equal(t, 1500000, *opts.MaxPrice)
	assert.NotNil(t, opts.MinPrice)
	assert.Equal(t, 500000, *opts.MinPrice)
	assert.NotNil(t, opts.MaxStops)
	assert.Equal(t, 1, *opts.MaxStops)
	assert.Len(t, opts.Airlines, 2)
	assert.Contains(t, opts.Airlines, "Garuda Indonesia")
	assert.Contains(t, opts.Airlines, "Lion Air")
}

func TestMultiCityLeg_Structure(t *testing.T) {
	leg := MultiCityLeg{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	assert.Equal(t, "CGK", leg.Origin)
	assert.Equal(t, "DPS", leg.Destination)
	assert.Equal(t, "2025-12-15", leg.DepartureDate)
}

func TestMultiCitySearchRequest_Structure(t *testing.T) {
	req := MultiCitySearchRequest{
		Legs: []MultiCityLeg{
			{Origin: "CGK", Destination: "DPS", DepartureDate: "2025-12-15"},
			{Origin: "DPS", Destination: "SUB", DepartureDate: "2025-12-20"},
		},
		Passengers: 2,
		CabinClass: "business",
	}

	assert.Len(t, req.Legs, 2)
	assert.Equal(t, 2, req.Passengers)
	assert.Equal(t, "business", req.CabinClass)
}
