package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/transport"
)

// GarudaProvider implements the Provider interface for Garuda Indonesia
type GarudaProvider struct {
	transport transport.Transport
	name      string
}

// NewGarudaProvider creates a new Garuda Indonesia provider with the given transport
func NewGarudaProvider(transport transport.Transport) *GarudaProvider {
	return &GarudaProvider{
		transport: transport,
		name:      "Garuda Indonesia",
	}
}

func (p *GarudaProvider) Name() string {
	return p.name
}

func (p *GarudaProvider) Search(ctx context.Context, request models.SearchRequest) ([]models.UnifiedFlight, error) {
	// Use transport to fetch raw data
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

	// Parse the raw response using provider-specific logic
	return p.parseResponse(resp.Body, request.Origin, request.Destination)
}

// parseResponse parses Garuda Indonesia specific JSON format
func (p *GarudaProvider) parseResponse(data []byte, origin, destination string) ([]models.UnifiedFlight, error) {
	var response GarudaResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	var flights []models.UnifiedFlight
	for _, f := range response.Flights {
		if f.Departure.Airport != origin || f.Arrival.Airport != destination {
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

// GarudaResponse represents the raw response from Garuda Indonesia API
type GarudaResponse struct {
	Status  string         `json:"status"`
	Flights []GarudaFlight `json:"flights"`
}

type GarudaFlight struct {
	FlightID        string          `json:"flight_id"`
	Airline         string          `json:"airline"`
	AirlineCode     string          `json:"airline_code"`
	Departure       GarudaLocation  `json:"departure"`
	Arrival         GarudaLocation  `json:"arrival"`
	DurationMinutes int             `json:"duration_minutes"`
	Stops           int             `json:"stops"`
	Aircraft        string          `json:"aircraft"`
	Price           GarudaPrice     `json:"price"`
	AvailableSeats  int             `json:"available_seats"`
	FareClass       string          `json:"fare_class"`
	Baggage         GarudaBaggage   `json:"baggage"`
	Amenities       []string        `json:"amenities"`
	Segments        []GarudaSegment `json:"segments,omitempty"`
}

type GarudaLocation struct {
	Airport  string `json:"airport"`
	City     string `json:"city"`
	Time     string `json:"time"`
	Terminal string `json:"terminal"`
}

type GarudaPrice struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

type GarudaBaggage struct {
	CarryOn int `json:"carry_on"`
	Checked int `json:"checked"`
}

type GarudaSegment struct {
	FlightNumber    string      `json:"flight_number"`
	Departure       GarudaPoint `json:"departure"`
	Arrival         GarudaPoint `json:"arrival"`
	DurationMinutes int         `json:"duration_minutes"`
	LayoverMinutes  *int        `json:"layover_minutes,omitempty"`
}

type GarudaPoint struct {
	Airport string `json:"airport"`
	Time    string `json:"time"`
}

func (p *GarudaProvider) convertToUnified(f GarudaFlight) (models.UnifiedFlight, error) {
	departureTime, err := time.Parse(time.RFC3339, f.Departure.Time)
	if err != nil {
		return models.UnifiedFlight{}, fmt.Errorf("failed to parse departure time: %w", err)
	}
	arrivalTime, err := time.Parse(time.RFC3339, f.Arrival.Time)
	if err != nil {
		return models.UnifiedFlight{}, fmt.Errorf("failed to parse arrival time: %w", err)
	}

	flight := models.UnifiedFlight{
		ID:           f.FlightID + "_Garuda Indonesia",
		Provider:     p.name,
		FlightNumber: f.FlightID,
		Airline: models.AirlineInfo{
			Name: f.Airline,
			Code: f.AirlineCode,
		},
		Departure: models.DepartureInfo{
			Airport:   f.Departure.Airport,
			City:      f.Departure.City,
			Datetime:  f.Departure.Time,
			Timestamp: departureTime.Unix(),
			Terminal:  f.Departure.Terminal,
		},
		Arrival: models.ArrivalInfo{
			Airport:   f.Arrival.Airport,
			City:      f.Arrival.City,
			Datetime:  f.Arrival.Time,
			Timestamp: arrivalTime.Unix(),
			Terminal:  f.Arrival.Terminal,
		},
		Duration: models.DurationInfo{
			TotalMinutes: f.DurationMinutes,
			Formatted:    models.FormatDuration(f.DurationMinutes),
		},
		Stops:          f.Stops,
		Price:          models.PriceInfo{Amount: f.Price.Amount, Currency: f.Price.Currency},
		AvailableSeats: f.AvailableSeats,
		CabinClass:     f.FareClass,
		Aircraft:       &f.Aircraft,
		Amenities:      f.Amenities,
		Baggage: models.BaggageInfo{
			CarryOn: formatBaggageCount(f.Baggage.CarryOn),
			Checked: formatBaggageCount(f.Baggage.Checked),
		},
	}

	if len(f.Segments) > 0 {
		flight.Segments = make([]models.FlightSegment, len(f.Segments))
		for i, seg := range f.Segments {
			depTime, err := time.Parse(time.RFC3339, seg.Departure.Time)
			if err != nil {
				return models.UnifiedFlight{}, fmt.Errorf("failed to parse segment departure time: %w", err)
			}
			arrTime, err := time.Parse(time.RFC3339, seg.Arrival.Time)
			if err != nil {
				return models.UnifiedFlight{}, fmt.Errorf("failed to parse segment arrival time: %w", err)
			}

			flight.Segments[i] = models.FlightSegment{
				FlightNumber:    seg.FlightNumber,
				DurationMinutes: seg.DurationMinutes,
				LayoverMinutes:  seg.LayoverMinutes,
				Departure: models.SegmentPoint{
					Airport:   seg.Departure.Airport,
					Time:      seg.Departure.Time,
					Timestamp: depTime.Unix(),
				},
				Arrival: models.SegmentPoint{
					Airport:   seg.Arrival.Airport,
					Time:      seg.Arrival.Time,
					Timestamp: arrTime.Unix(),
				},
			}
		}
	}

	return flight, nil
}

func formatBaggageCount(count int) string {
	if count == 0 {
		return "Not included"
	}
	if count == 1 {
		return "1 piece"
	}
	return "2 pieces"
}
