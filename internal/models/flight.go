package models

import (
	"fmt"
	"time"
)

// SearchRequest represents the incoming search request
type SearchRequest struct {
	Origin        string  `json:"origin"`
	Destination   string  `json:"destination"`
	DepartureDate string  `json:"departureDate"`
	ReturnDate    *string `json:"returnDate,omitempty"`
	Passengers    int     `json:"passengers"`
	CabinClass    string  `json:"cabinClass"`
}

// MultiCityLeg represents a single leg in a multi-city journey
type MultiCityLeg struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
}

// MultiCitySearchRequest represents a multi-city search request
type MultiCitySearchRequest struct {
	Legs       []MultiCityLeg `json:"legs"`
	Passengers int            `json:"passengers"`
	CabinClass string         `json:"cabinClass"`
}

// SearchResponse represents the unified search response
type SearchResponse struct {
	SearchCriteria SearchCriteria  `json:"search_criteria"`
	Metadata       Metadata        `json:"metadata"`
	Flights        []UnifiedFlight `json:"flights"`
}

// SearchCriteria mirrors the search request
type SearchCriteria struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabin_class"`
}

// Metadata contains search execution information
type Metadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

// UnifiedFlight represents a normalized flight across all providers
type UnifiedFlight struct {
	ID             string          `json:"id"`
	Provider       string          `json:"provider"`
	Airline        AirlineInfo     `json:"airline"`
	FlightNumber   string          `json:"flight_number"`
	Departure      DepartureInfo   `json:"departure"`
	Arrival        ArrivalInfo     `json:"arrival"`
	Duration       DurationInfo    `json:"duration"`
	Stops          int             `json:"stops"`
	Price          PriceInfo       `json:"price"`
	AvailableSeats int             `json:"available_seats"`
	CabinClass     string          `json:"cabin_class"`
	Aircraft       *string         `json:"aircraft,omitempty"`
	Amenities      []string        `json:"amenities"`
	Baggage        BaggageInfo     `json:"baggage"`
	Segments       []FlightSegment `json:"segments,omitempty"`
	BestValueScore float64         `json:"best_value_score,omitempty"`
}

// AirlineInfo contains airline details
type AirlineInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// DepartureInfo contains departure details
type DepartureInfo struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	Datetime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
	Terminal  string `json:"terminal,omitempty"`
}

// ArrivalInfo contains arrival details
type ArrivalInfo struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	Datetime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
	Terminal  string `json:"terminal,omitempty"`
}

// DurationInfo contains duration details
type DurationInfo struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

// PriceInfo contains price details
type PriceInfo struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

// BaggageInfo contains baggage allowance details
type BaggageInfo struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}

// FlightSegment represents a segment of a flight with layovers
type FlightSegment struct {
	FlightNumber    string       `json:"flight_number"`
	Departure       SegmentPoint `json:"departure"`
	Arrival         SegmentPoint `json:"arrival"`
	DurationMinutes int          `json:"duration_minutes"`
	LayoverMinutes  *int         `json:"layover_minutes,omitempty"`
}

// SegmentPoint represents a departure or arrival point in a segment
type SegmentPoint struct {
	Airport   string `json:"airport"`
	Time      string `json:"time"`
	Timestamp int64  `json:"timestamp"`
}

// FilterOptions contains available filter parameters
type FilterOptions struct {
	MaxPrice        *int       `json:"max_price,omitempty"`
	MinPrice        *int       `json:"min_price,omitempty"`
	MaxStops        *int       `json:"max_stops,omitempty"`
	Airlines        []string   `json:"airlines,omitempty"`
	DepartureAfter  *time.Time `json:"departure_after,omitempty"`
	DepartureBefore *time.Time `json:"departure_before,omitempty"`
	ArrivalAfter    *time.Time `json:"arrival_after,omitempty"`
	ArrivalBefore   *time.Time `json:"arrival_before,omitempty"`
	MaxDurationMins *int       `json:"max_duration_mins,omitempty"`
}

// SortOption represents sorting options
type SortOption string

const (
	SortByPriceAsc      SortOption = "price_asc"
	SortByPriceDesc     SortOption = "price_desc"
	SortByDurationAsc   SortOption = "duration_asc"
	SortByDurationDesc  SortOption = "duration_desc"
	SortByDepartureAsc  SortOption = "departure_asc"
	SortByDepartureDesc SortOption = "departure_desc"
	SortByArrivalAsc    SortOption = "arrival_asc"
	SortByArrivalDesc   SortOption = "arrival_desc"
	SortByBestValue     SortOption = "best_value"
)

// FormatDuration formats minutes to human readable string
func FormatDuration(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60
	if hours > 0 && mins > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", mins)
}

// ParseDurationString parses duration strings like "1h 45m" or "3h 5m" to minutes
func ParseDurationString(duration string) int {
	var hours, minutes int
	fmt.Sscanf(duration, "%dh %dm", &hours, &minutes)
	return hours*60 + minutes
}

// FormatPrice formats price with thousands separator for IDR
func FormatPrice(amount int) string {
	if amount < 1000 {
		return fmt.Sprintf("%d", amount)
	}

	result := ""
	str := fmt.Sprintf("%d", amount)
	length := len(str)

	for i, c := range str {
		if i > 0 && (length-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}

	return result
}
