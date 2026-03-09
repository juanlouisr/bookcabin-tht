package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/transport"
)

// BatikAirProvider implements the Provider interface for Batik Air
type BatikAirProvider struct {
	transport transport.Transport
	name      string
}

// NewBatikAirProvider creates a new Batik Air provider with the given transport
func NewBatikAirProvider(transport transport.Transport) *BatikAirProvider {
	return &BatikAirProvider{
		transport: transport,
		name:      "Batik Air",
	}
}

func (p *BatikAirProvider) Name() string {
	return p.name
}

func (p *BatikAirProvider) Search(ctx context.Context, request models.SearchRequest) ([]models.UnifiedFlight, error) {
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

// parseResponse parses Batik Air specific JSON format
func (p *BatikAirProvider) parseResponse(data []byte, origin, destination string) ([]models.UnifiedFlight, error) {
	var response BatikAirResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Batik Air response: %w", err)
	}

	// Validate response code
	if response.Code != 200 {
		return nil, fmt.Errorf("batik air API returned non-200 code: %d, message: %s", response.Code, response.Message)
	}

	var flights []models.UnifiedFlight
	for _, f := range response.Results {
		if f.Origin != origin || f.Destination != destination {
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

// BatikAirResponse represents the raw response from Batik Air API
type BatikAirResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Results []BatikAirFlight `json:"results"`
}

type BatikAirFlight struct {
	FlightNumber      string               `json:"flightNumber"`
	AirlineName       string               `json:"airlineName"`
	AirlineIATA       string               `json:"airlineIATA"`
	Origin            string               `json:"origin"`
	Destination       string               `json:"destination"`
	DepartureDateTime string               `json:"departureDateTime"`
	ArrivalDateTime   string               `json:"arrivalDateTime"`
	TravelTime        string               `json:"travelTime"`
	NumberOfStops     int                  `json:"numberOfStops"`
	Connections       []BatikAirConnection `json:"connections,omitempty"`
	Fare              BatikAirFare         `json:"fare"`
	SeatsAvailable    int                  `json:"seatsAvailable"`
	AircraftModel     string               `json:"aircraftModel"`
	BaggageInfo       string               `json:"baggageInfo"`
	OnboardServices   []string             `json:"onboardServices"`
}

type BatikAirConnection struct {
	StopAirport  string `json:"stopAirport"`
	StopDuration string `json:"stopDuration"`
}

type BatikAirFare struct {
	BasePrice    int    `json:"basePrice"`
	Taxes        int    `json:"taxes"`
	TotalPrice   int    `json:"totalPrice"`
	CurrencyCode string `json:"currencyCode"`
	Class        string `json:"class"`
}

func (p *BatikAirProvider) convertToUnified(f BatikAirFlight) (models.UnifiedFlight, error) {
	departureTime, err := time.Parse("2006-01-02T15:04:05-0700", f.DepartureDateTime)
	if err != nil {
		return models.UnifiedFlight{}, fmt.Errorf("failed to parse departure time: %w", err)
	}
	arrivalTime, err := time.Parse("2006-01-02T15:04:05-0700", f.ArrivalDateTime)
	if err != nil {
		return models.UnifiedFlight{}, fmt.Errorf("failed to parse arrival time: %w", err)
	}

	durationMinutes := models.ParseDurationString(f.TravelTime)

	return models.UnifiedFlight{
		ID:           f.FlightNumber + "_Batik Air",
		Provider:     p.name,
		FlightNumber: f.FlightNumber,
		Airline: models.AirlineInfo{
			Name: f.AirlineName,
			Code: f.AirlineIATA,
		},
		Departure: models.DepartureInfo{
			Airport:   f.Origin,
			City:      getCityFromAirport(f.Origin),
			Datetime:  departureTime.Format(time.RFC3339),
			Timestamp: departureTime.Unix(),
		},
		Arrival: models.ArrivalInfo{
			Airport:   f.Destination,
			City:      getCityFromAirport(f.Destination),
			Datetime:  arrivalTime.Format(time.RFC3339),
			Timestamp: arrivalTime.Unix(),
		},
		Duration: models.DurationInfo{
			TotalMinutes: durationMinutes,
			Formatted:    f.TravelTime,
		},
		Stops:          f.NumberOfStops,
		Price:          models.PriceInfo{Amount: f.Fare.TotalPrice, Currency: f.Fare.CurrencyCode},
		AvailableSeats: f.SeatsAvailable,
		CabinClass:     f.Fare.Class,
		Aircraft:       &f.AircraftModel,
		Amenities:      f.OnboardServices,
		Baggage: models.BaggageInfo{
			CarryOn: "7kg",
			Checked: "20kg",
		},
	}, nil
}

// ParseBatikAirResponse parses raw Batik Air JSON data (exported for testing)
func ParseBatikAirResponse(data []byte) (*BatikAirResponse, error) {
	var response BatikAirResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Batik Air response: %w", err)
	}
	return &response, nil
}
