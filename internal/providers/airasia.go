package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/transport"
)

// AirAsiaProvider implements the Provider interface for AirAsia
type AirAsiaProvider struct {
	transport transport.Transport
	name      string
}

// NewAirAsiaProvider creates a new AirAsia provider with the given transport
func NewAirAsiaProvider(transport transport.Transport) *AirAsiaProvider {
	return &AirAsiaProvider{
		transport: transport,
		name:      "AirAsia",
	}
}

func (p *AirAsiaProvider) Name() string {
	return p.name
}

func (p *AirAsiaProvider) Search(ctx context.Context, request models.SearchRequest) ([]models.UnifiedFlight, error) {
	transportReq := transport.Request{
		Origin:        request.Origin,
		Destination:   request.Destination,
		DepartureDate: request.DepartureDate,
		ReturnDate:    request.ReturnDate,
		Passengers:    request.Passengers,
		CabinClass:    request.CabinClass,
	}

	resp, err := p.transport.Fetch(ctx, transportReq)
	if err != nil {
		return nil, err
	}

	return p.parseResponse(resp.Body, request.Origin, request.Destination)
}

// parseResponse parses AirAsia specific JSON format
func (p *AirAsiaProvider) parseResponse(data []byte, origin, destination string) ([]models.UnifiedFlight, error) {
	var response AirAsiaResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AirAsia response: %w", err)
	}

	// Validate response status
	if response.Status != "ok" {
		return nil, fmt.Errorf("airasia API returned non-ok status: %s", response.Status)
	}

	var flights []models.UnifiedFlight
	for _, f := range response.Flights {
		if f.FromAirport != origin || f.ToAirport != destination {
			continue
		}

		flight, err := p.convertToUnified(f)
		if err != nil {
			// Log error but continue processing other flights
			continue
		}
		flights = append(flights, flight)
	}

	return flights, nil
}

// AirAsiaResponse represents the raw response from AirAsia API
type AirAsiaResponse struct {
	Status  string          `json:"status"`
	Flights []AirAsiaFlight `json:"flights"`
}

type AirAsiaFlight struct {
	FlightCode    string        `json:"flight_code"`
	Airline       string        `json:"airline"`
	FromAirport   string        `json:"from_airport"`
	ToAirport     string        `json:"to_airport"`
	DepartTime    string        `json:"depart_time"`
	ArriveTime    string        `json:"arrive_time"`
	DurationHours float64       `json:"duration_hours"`
	DirectFlight  bool          `json:"direct_flight"`
	Stops         []AirAsiaStop `json:"stops,omitempty"`
	PriceIDR      int           `json:"price_idr"`
	Seats         int           `json:"seats"`
	CabinClass    string        `json:"cabin_class"`
	BaggageNote   string        `json:"baggage_note"`
}

type AirAsiaStop struct {
	Airport         string `json:"airport"`
	WaitTimeMinutes int    `json:"wait_time_minutes"`
}

func (p *AirAsiaProvider) convertToUnified(f AirAsiaFlight) (models.UnifiedFlight, error) {
	departureTime, err := time.Parse(time.RFC3339, f.DepartTime)
	if err != nil {
		return models.UnifiedFlight{}, fmt.Errorf("failed to parse departure time: %w", err)
	}
	arrivalTime, err := time.Parse(time.RFC3339, f.ArriveTime)
	if err != nil {
		return models.UnifiedFlight{}, fmt.Errorf("failed to parse arrival time: %w", err)
	}

	stops := 0
	if !f.DirectFlight {
		stops = len(f.Stops)
	}

	durationMinutes := int(f.DurationHours * 60)

	return models.UnifiedFlight{
		ID:           f.FlightCode + "_AirAsia",
		Provider:     p.name,
		FlightNumber: f.FlightCode,
		Airline: models.AirlineInfo{
			Name: f.Airline,
			Code: "QZ",
		},
		Departure: models.DepartureInfo{
			Airport:   f.FromAirport,
			City:      getCityFromAirport(f.FromAirport),
			Datetime:  f.DepartTime,
			Timestamp: departureTime.Unix(),
		},
		Arrival: models.ArrivalInfo{
			Airport:   f.ToAirport,
			City:      getCityFromAirport(f.ToAirport),
			Datetime:  f.ArriveTime,
			Timestamp: arrivalTime.Unix(),
		},
		Duration: models.DurationInfo{
			TotalMinutes: durationMinutes,
			Formatted:    models.FormatDuration(durationMinutes),
		},
		Stops:          stops,
		Price:          models.PriceInfo{Amount: f.PriceIDR, Currency: "IDR"},
		AvailableSeats: f.Seats,
		CabinClass:     f.CabinClass,
		Amenities:      []string{},
		Baggage: models.BaggageInfo{
			CarryOn: "Cabin baggage only",
			Checked: "Additional fee",
		},
	}, nil
}

// ParseAirAsiaResponse parses raw AirAsia JSON data (exported for testing)
func ParseAirAsiaResponse(data []byte) (*AirAsiaResponse, error) {
	var response AirAsiaResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AirAsia response: %w", err)
	}
	return &response, nil
}

// ProviderError represents an error from a provider
type ProviderError struct {
	Provider string
	Message  string
}

func (e *ProviderError) Error() string {
	return e.Provider + ": " + e.Message
}
