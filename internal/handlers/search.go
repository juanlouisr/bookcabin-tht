package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bookcabin/flight/internal/aggregator"
	"bookcabin/flight/internal/filters"
	"bookcabin/flight/internal/models"
)

// SearchHandler handles flight search requests
type SearchHandler struct {
	aggregator *aggregator.Service
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(agg *aggregator.Service) *SearchHandler {
	return &SearchHandler{aggregator: agg}
}

// SearchRequest represents the HTTP request body for flight search
type SearchHTTPRequest struct {
	Origin        string        `json:"origin"`
	Destination   string        `json:"destination"`
	DepartureDate string        `json:"departure_date"`
	ReturnDate    *string       `json:"return_date,omitempty"`
	Passengers    int           `json:"passengers"`
	CabinClass    string        `json:"cabin_class"`
	Filters       FilterOptions `json:"filters,omitempty"`
	SortBy        string        `json:"sort_by,omitempty"`
}

// FilterOptions represents filter parameters
type FilterOptions struct {
	MaxPrice        *int     `json:"max_price,omitempty"`
	MinPrice        *int     `json:"min_price,omitempty"`
	MaxStops        *int     `json:"max_stops,omitempty"`
	Airlines        []string `json:"airlines,omitempty"`
	DepartureAfter  *string  `json:"departure_after,omitempty"`
	DepartureBefore *string  `json:"departure_before,omitempty"`
	ArrivalAfter    *string  `json:"arrival_after,omitempty"`
	ArrivalBefore   *string  `json:"arrival_before,omitempty"`
	MaxDurationMins *int     `json:"max_duration_mins,omitempty"`
}

// HandleSearch handles the search endpoint
func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Origin == "" || req.Destination == "" || req.DepartureDate == "" {
		http.Error(w, "Missing required fields: origin, destination, departure_date", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Passengers == 0 {
		req.Passengers = 1
	}
	if req.CabinClass == "" {
		req.CabinClass = "economy"
	}

	// Create search request
	searchReq := models.SearchRequest{
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureDate: req.DepartureDate,
		ReturnDate:    req.ReturnDate,
		Passengers:    req.Passengers,
		CabinClass:    req.CabinClass,
	}

	// Execute search
	response := h.aggregator.Search(r.Context(), searchReq)

	// Apply filters
	filterOptions := convertFilterOptions(req.Filters)
	response.Flights = filters.ApplyFilters(response.Flights, filterOptions)

	// Apply sorting
	sortOption := convertSortOption(req.SortBy)
	response.Flights = filters.SortFlights(response.Flights, sortOption)

	// Update metadata
	response.Metadata.TotalResults = len(response.Flights)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// MultiCitySearchRequest represents the HTTP request body for multi-city search
type MultiCitySearchHTTPRequest struct {
	Legs       []models.MultiCityLeg `json:"legs"`
	Passengers int                   `json:"passengers"`
	CabinClass string                `json:"cabin_class"`
}

// HandleMultiCitySearch handles the multi-city search endpoint
func (h *SearchHandler) HandleMultiCitySearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MultiCitySearchHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if len(req.Legs) == 0 {
		http.Error(w, "Missing required field: legs (at least one leg required)", http.StatusBadRequest)
		return
	}

	// Validate each leg
	for i, leg := range req.Legs {
		if leg.Origin == "" || leg.Destination == "" || leg.DepartureDate == "" {
			http.Error(w, fmt.Sprintf("Missing required fields in leg %d: origin, destination, departure_date", i+1), http.StatusBadRequest)
			return
		}
	}

	// Set defaults
	if req.Passengers == 0 {
		req.Passengers = 1
	}
	if req.CabinClass == "" {
		req.CabinClass = "economy"
	}

	// Create multi-city search request
	searchReq := models.MultiCitySearchRequest{
		Legs:       req.Legs,
		Passengers: req.Passengers,
		CabinClass: req.CabinClass,
	}

	// Execute multi-city search
	responses := h.aggregator.SearchMultiCity(r.Context(), searchReq)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"legs": responses,
		"metadata": map[string]interface{}{
			"total_legs":  len(responses),
			"search_time": time.Now().Format(time.RFC3339),
		},
	})
}

// HandleHealth returns health check status
func (h *SearchHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// convertFilterOptions converts HTTP filter options to model filter options
func convertFilterOptions(opts FilterOptions) models.FilterOptions {
	result := models.FilterOptions{
		MaxPrice:        opts.MaxPrice,
		MinPrice:        opts.MinPrice,
		MaxStops:        opts.MaxStops,
		Airlines:        opts.Airlines,
		MaxDurationMins: opts.MaxDurationMins,
	}

	// Parse time strings
	if opts.DepartureAfter != nil {
		t, _ := time.Parse("15:04", *opts.DepartureAfter)
		result.DepartureAfter = &t
	}
	if opts.DepartureBefore != nil {
		t, _ := time.Parse("15:04", *opts.DepartureBefore)
		result.DepartureBefore = &t
	}
	if opts.ArrivalAfter != nil {
		t, _ := time.Parse("15:04", *opts.ArrivalAfter)
		result.ArrivalAfter = &t
	}
	if opts.ArrivalBefore != nil {
		t, _ := time.Parse("15:04", *opts.ArrivalBefore)
		result.ArrivalBefore = &t
	}

	return result
}

// convertSortOption converts sort string to SortOption
func convertSortOption(sortBy string) models.SortOption {
	switch strings.ToLower(sortBy) {
	case "price_asc", "price", "cheapest":
		return models.SortByPriceAsc
	case "price_desc":
		return models.SortByPriceDesc
	case "duration_asc", "duration", "fastest":
		return models.SortByDurationAsc
	case "duration_desc":
		return models.SortByDurationDesc
	case "departure_asc", "departure":
		return models.SortByDepartureAsc
	case "departure_desc":
		return models.SortByDepartureDesc
	case "arrival_asc", "arrival":
		return models.SortByArrivalAsc
	case "arrival_desc":
		return models.SortByArrivalDesc
	case "best_value", "bestvalue", "recommended":
		return models.SortByBestValue
	default:
		return models.SortByPriceAsc // Default sort
	}
}

// parseQueryParams parses filter options from query parameters (for GET requests)
func parseQueryParams(r *http.Request) FilterOptions {
	query := r.URL.Query()
	opts := FilterOptions{}

	if v := query.Get("max_price"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			opts.MaxPrice = &i
		}
	}

	if v := query.Get("min_price"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			opts.MinPrice = &i
		}
	}

	if v := query.Get("max_stops"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			opts.MaxStops = &i
		}
	}

	if v := query.Get("airlines"); v != "" {
		opts.Airlines = strings.Split(v, ",")
	}

	if v := query.Get("max_duration_mins"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			opts.MaxDurationMins = &i
		}
	}

	return opts
}
