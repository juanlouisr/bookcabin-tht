package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/transport"
)

// LionAirProvider implements the Provider interface for Lion Air
type LionAirProvider struct {
	transport transport.Transport
	name      string
}

// NewLionAirProvider creates a new Lion Air provider with the given transport
func NewLionAirProvider(transport transport.Transport) *LionAirProvider {
	return &LionAirProvider{
		transport: transport,
		name:      "Lion Air",
	}
}

func (p *LionAirProvider) Name() string {
	return p.name
}

func (p *LionAirProvider) Search(ctx context.Context, request models.SearchRequest) ([]models.UnifiedFlight, error) {
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

// parseResponse parses Lion Air specific JSON format
func (p *LionAirProvider) parseResponse(data []byte, origin, destination string) ([]models.UnifiedFlight, error) {
	var response LionAirResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Lion Air response: %w", err)
	}

	// Validate response
	if !response.Success {
		return nil, fmt.Errorf("lion air API returned unsuccessful response")
	}

	var flights []models.UnifiedFlight
	for _, f := range response.Data.AvailableFlights {
		if f.Route.From.Code != origin || f.Route.To.Code != destination {
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

// LionAirResponse represents the raw response from Lion Air API
type LionAirResponse struct {
	Success bool        `json:"success"`
	Data    LionAirData `json:"data"`
}

type LionAirData struct {
	AvailableFlights []LionAirFlight `json:"available_flights"`
}

type LionAirFlight struct {
	ID         string           `json:"id"`
	Carrier    LionAirCarrier   `json:"carrier"`
	Route      LionAirRoute     `json:"route"`
	Schedule   LionAirSchedule  `json:"schedule"`
	FlightTime int              `json:"flight_time"`
	IsDirect   bool             `json:"is_direct"`
	StopCount  int              `json:"stop_count,omitempty"`
	Layovers   []LionAirLayover `json:"layovers,omitempty"`
	Pricing    LionAirPricing   `json:"pricing"`
	SeatsLeft  int              `json:"seats_left"`
	PlaneType  string           `json:"plane_type"`
	Services   LionAirServices  `json:"services"`
}

type LionAirCarrier struct {
	Name string `json:"name"`
	IATA string `json:"iata"`
}

type LionAirRoute struct {
	From LionAirAirport `json:"from"`
	To   LionAirAirport `json:"to"`
}

type LionAirAirport struct {
	Code string `json:"code"`
	Name string `json:"name"`
	City string `json:"city"`
}

type LionAirSchedule struct {
	Departure         string `json:"departure"`
	DepartureTimezone string `json:"departure_timezone"`
	Arrival           string `json:"arrival"`
	ArrivalTimezone   string `json:"arrival_timezone"`
}

type LionAirLayover struct {
	Airport         string `json:"airport"`
	DurationMinutes int    `json:"duration_minutes"`
}

type LionAirPricing struct {
	Total    int    `json:"total"`
	Currency string `json:"currency"`
	FareType string `json:"fare_type"`
}

type LionAirServices struct {
	WifiAvailable    bool           `json:"wifi_available"`
	MealsIncluded    bool           `json:"meals_included"`
	BaggageAllowance LionAirBaggage `json:"baggage_allowance"`
}

type LionAirBaggage struct {
	Cabin string `json:"cabin"`
	Hold  string `json:"hold"`
}

func (p *LionAirProvider) convertToUnified(f LionAirFlight) (models.UnifiedFlight, error) {
	departureTime, err := time.Parse("2006-01-02T15:04:05", f.Schedule.Departure)
	if err != nil {
		return models.UnifiedFlight{}, fmt.Errorf("failed to parse departure time: %w", err)
	}
	arrivalTime, err := time.Parse("2006-01-02T15:04:05", f.Schedule.Arrival)
	if err != nil {
		return models.UnifiedFlight{}, fmt.Errorf("failed to parse arrival time: %w", err)
	}

	depLocation, err := time.LoadLocation(f.Schedule.DepartureTimezone)
	if err != nil {
		depLocation = time.UTC
	}
	arrLocation, err := time.LoadLocation(f.Schedule.ArrivalTimezone)
	if err != nil {
		arrLocation = time.UTC
	}

	departureTime = time.Date(
		departureTime.Year(), departureTime.Month(), departureTime.Day(),
		departureTime.Hour(), departureTime.Minute(), 0, 0, depLocation,
	)
	arrivalTime = time.Date(
		arrivalTime.Year(), arrivalTime.Month(), arrivalTime.Day(),
		arrivalTime.Hour(), arrivalTime.Minute(), 0, 0, arrLocation,
	)

	stops := 0
	if !f.IsDirect {
		stops = f.StopCount
		if stops == 0 {
			stops = len(f.Layovers)
		}
	}

	var amenities []string
	if f.Services.WifiAvailable {
		amenities = append(amenities, "wifi")
	}
	if f.Services.MealsIncluded {
		amenities = append(amenities, "meal")
	}

	return models.UnifiedFlight{
		ID:           f.ID + "_Lion Air",
		Provider:     p.name,
		FlightNumber: f.ID,
		Airline: models.AirlineInfo{
			Name: f.Carrier.Name,
			Code: f.Carrier.IATA,
		},
		Departure: models.DepartureInfo{
			Airport:   f.Route.From.Code,
			City:      f.Route.From.City,
			Datetime:  departureTime.Format(time.RFC3339),
			Timestamp: departureTime.Unix(),
		},
		Arrival: models.ArrivalInfo{
			Airport:   f.Route.To.Code,
			City:      f.Route.To.City,
			Datetime:  arrivalTime.Format(time.RFC3339),
			Timestamp: arrivalTime.Unix(),
		},
		Duration: models.DurationInfo{
			TotalMinutes: f.FlightTime,
			Formatted:    models.FormatDuration(f.FlightTime),
		},
		Stops:          stops,
		Price:          models.PriceInfo{Amount: f.Pricing.Total, Currency: f.Pricing.Currency},
		AvailableSeats: f.SeatsLeft,
		CabinClass:     f.Pricing.FareType,
		Aircraft:       &f.PlaneType,
		Amenities:      amenities,
		Baggage: models.BaggageInfo{
			CarryOn: f.Services.BaggageAllowance.Cabin,
			Checked: f.Services.BaggageAllowance.Hold,
		},
	}, nil
}

// ParseLionAirResponse parses raw Lion Air JSON data (exported for testing)
func ParseLionAirResponse(data []byte) (*LionAirResponse, error) {
	var response LionAirResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Lion Air response: %w", err)
	}
	return &response, nil
}
