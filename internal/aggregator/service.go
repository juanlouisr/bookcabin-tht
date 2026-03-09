package aggregator

import (
	"context"
	"sync"
	"time"

	"bookcabin/flight/internal/cache"
	"bookcabin/flight/internal/models"
)

// Service aggregates flight data from multiple providers
type Service struct {
	providers []models.Provider
	cache     cache.Cache
	timeout   time.Duration
}

// NewService creates a new aggregation service
func NewService(providers []models.Provider, cache cache.Cache, timeout time.Duration) *Service {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Service{
		providers: providers,
		cache:     cache,
		timeout:   timeout,
	}
}

// Search aggregates flight results from all providers
func (s *Service) Search(ctx context.Context, request models.SearchRequest) models.SearchResponse {
	startTime := time.Now()

	// Create search criteria for response
	criteria := models.SearchCriteria{
		Origin:        request.Origin,
		Destination:   request.Destination,
		DepartureDate: request.DepartureDate,
		Passengers:    request.Passengers,
		CabinClass:    request.CabinClass,
	}

	// Check cache first
	cacheKey := generateCacheKey(request)
	if s.cache != nil {
		if cached, found := s.cache.Get(cacheKey); found {
			response := cached.(models.SearchResponse)
			response.Metadata.CacheHit = true
			response.Metadata.SearchTimeMs = time.Since(startTime).Milliseconds()
			return response
		}
	}

	// Create context with timeout
	searchCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Query all providers concurrently
	results := s.queryProviders(searchCtx, request)

	// Aggregate results
	var allFlights []models.UnifiedFlight
	successCount := 0
	failCount := 0

	for _, result := range results {
		if result.Error != nil {
			failCount++
			continue
		}
		successCount++
		allFlights = append(allFlights, result.Flights...)
	}

	// Calculate best value scores
	allFlights = calculateBestValueScores(allFlights)

	// Build response
	response := models.SearchResponse{
		SearchCriteria: criteria,
		Metadata: models.Metadata{
			TotalResults:       len(allFlights),
			ProvidersQueried:   len(s.providers),
			ProvidersSucceeded: successCount,
			ProvidersFailed:    failCount,
			SearchTimeMs:       time.Since(startTime).Milliseconds(),
			CacheHit:           false,
		},
		Flights: allFlights,
	}

	// Cache the result
	if s.cache != nil {
		s.cache.Set(cacheKey, response, 5*time.Minute)
	}

	return response
}

// SearchMultiCity performs a multi-city search across all providers
func (s *Service) SearchMultiCity(ctx context.Context, request models.MultiCitySearchRequest) []models.SearchResponse {
	startTime := time.Now()

	if len(request.Legs) == 0 {
		return []models.SearchResponse{}
	}

	// Search each leg independently
	responses := make([]models.SearchResponse, len(request.Legs))
	var totalResults int
	var totalProvidersQueried, totalProvidersSucceeded, totalProvidersFailed int

	for i, leg := range request.Legs {
		searchReq := models.SearchRequest{
			Origin:        leg.Origin,
			Destination:   leg.Destination,
			DepartureDate: leg.DepartureDate,
			Passengers:    request.Passengers,
			CabinClass:    request.CabinClass,
		}

		responses[i] = s.Search(ctx, searchReq)
		totalResults += responses[i].Metadata.TotalResults
		totalProvidersQueried += responses[i].Metadata.ProvidersQueried
		totalProvidersSucceeded += responses[i].Metadata.ProvidersSucceeded
		totalProvidersFailed += responses[i].Metadata.ProvidersFailed
	}

	// Update metadata for all responses
	searchTimeMs := time.Since(startTime).Milliseconds()
	for i := range responses {
		responses[i].Metadata.TotalResults = totalResults
		responses[i].Metadata.ProvidersQueried = totalProvidersQueried
		responses[i].Metadata.ProvidersSucceeded = totalProvidersSucceeded
		responses[i].Metadata.ProvidersFailed = totalProvidersFailed
		responses[i].Metadata.SearchTimeMs = searchTimeMs
	}

	return responses
}

// queryProviders queries all providers concurrently
func (s *Service) queryProviders(ctx context.Context, request models.SearchRequest) []models.ProviderResult {
	var wg sync.WaitGroup
	results := make([]models.ProviderResult, len(s.providers))
	resultChan := make(chan struct {
		index  int
		result models.ProviderResult
	}, len(s.providers))

	// Query each provider concurrently
	for i, provider := range s.providers {
		wg.Add(1)
		go func(index int, p models.Provider) {
			defer wg.Done()

			flights, err := p.Search(ctx, request)
			resultChan <- struct {
				index  int
				result models.ProviderResult
			}{
				index: index,
				result: models.ProviderResult{
					Provider: p.Name(),
					Flights:  flights,
					Error:    err,
				},
			}
		}(i, provider)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for r := range resultChan {
		results[r.index] = r.result
	}

	return results
}

// generateCacheKey creates a cache key from the search request
func generateCacheKey(request models.SearchRequest) string {
	return request.Origin + "_" + request.Destination + "_" + request.DepartureDate + "_" + request.CabinClass
}

// calculateBestValueScores calculates a best value score for each flight
// Lower score is better (combination of price and duration)
func calculateBestValueScores(flights []models.UnifiedFlight) []models.UnifiedFlight {
	if len(flights) == 0 {
		return flights
	}

	// Find min/max values for normalization
	minPrice := flights[0].Price.Amount
	maxPrice := flights[0].Price.Amount
	minDuration := flights[0].Duration.TotalMinutes
	maxDuration := flights[0].Duration.TotalMinutes
	minStops := flights[0].Stops
	maxStops := flights[0].Stops

	for _, f := range flights {
		if f.Price.Amount < minPrice {
			minPrice = f.Price.Amount
		}
		if f.Price.Amount > maxPrice {
			maxPrice = f.Price.Amount
		}
		if f.Duration.TotalMinutes < minDuration {
			minDuration = f.Duration.TotalMinutes
		}
		if f.Duration.TotalMinutes > maxDuration {
			maxDuration = f.Duration.TotalMinutes
		}
		if f.Stops < minStops {
			minStops = f.Stops
		}
		if f.Stops > maxStops {
			maxStops = f.Stops
		}
	}

	// Avoid division by zero
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}
	durationRange := maxDuration - minDuration
	if durationRange == 0 {
		durationRange = 1
	}
	stopsRange := maxStops - minStops
	if stopsRange == 0 {
		stopsRange = 1
	}

	// Calculate score for each flight
	for i := range flights {
		// Normalize values (0-1 scale)
		normalizedPrice := float64(flights[i].Price.Amount-minPrice) / float64(priceRange)
		normalizedDuration := float64(flights[i].Duration.TotalMinutes-minDuration) / float64(durationRange)
		normalizedStops := float64(flights[i].Stops-minStops) / float64(stopsRange)

		// Weighted score: 50% price, 30% duration, 20% stops
		// Lower score is better
		flights[i].BestValueScore = normalizedPrice*0.5 + normalizedDuration*0.3 + normalizedStops*0.2
	}

	return flights
}
