package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bookcabin/flight/internal/aggregator"
	"bookcabin/flight/internal/cache"
	"bookcabin/flight/internal/handlers"
	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/providers"
)

func main() {
	// Initialize providers
	providerList, err := providers.CreateProviders()
	if err != nil {
		log.Fatalf("Failed to create providers: %v", err)
	}

	// Initialize cache
	flightCache := cache.NewMemoryCache()

	// Initialize aggregator service with 5 second timeout
	aggService := aggregator.NewService(providerList, flightCache, 5*time.Second)

	// Initialize handlers
	searchHandler := handlers.NewSearchHandler(aggService)
	swaggerHandler := handlers.NewSwaggerUIHandler("api/openapi.yaml")

	// Setup routes
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/search", searchHandler.HandleSearch)
	mux.HandleFunc("/api/v1/search/multi-city", searchHandler.HandleMultiCitySearch)
	mux.HandleFunc("/health", searchHandler.HandleHealth)

	// Swagger UI routes
	mux.HandleFunc("/api/docs", swaggerHandler.HandleSwaggerUI)
	mux.HandleFunc("/api/docs/", swaggerHandler.HandleSwaggerUI)
	mux.HandleFunc("/api/docs/openapi.yaml", swaggerHandler.HandleOpenAPISpec)

	// Create server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting Flight Search API server on port %s", port)
		log.Printf("Swagger UI available at: http://localhost:%s/api/docs", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Run demo search
	fmt.Println("=== Running Demo Search ===")
	runDemoSearch(aggService)
	fmt.Println("=== End of Demo ===")
	fmt.Println()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// runDemoSearch runs a sample search and displays results
func runDemoSearch(service *aggregator.Service) {
	request := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	fmt.Printf("Searching flights from %s to %s on %s...\n\n",
		request.Origin, request.Destination, request.DepartureDate)

	response := service.Search(context.Background(), request)

	// Print metadata
	fmt.Printf("Search Results:\n")
	fmt.Printf("- Total flights found: %d\n", response.Metadata.TotalResults)
	fmt.Printf("- Providers queried: %d\n", response.Metadata.ProvidersQueried)
	fmt.Printf("- Providers succeeded: %d\n", response.Metadata.ProvidersSucceeded)
	fmt.Printf("- Providers failed: %d\n", response.Metadata.ProvidersFailed)
	fmt.Printf("- Search time: %d ms\n", response.Metadata.SearchTimeMs)
	fmt.Printf("- Cache hit: %v\n\n", response.Metadata.CacheHit)

	// Print flights (limit to 5)
	fmt.Println("Top 5 Flights (sorted by price):")
	fmt.Println("----------------------------------")

	for i, flight := range response.Flights {
		if i >= 5 {
			break
		}

		fmt.Printf("\n%d. %s (%s)\n", i+1, flight.Airline.Name, flight.FlightNumber)
		fmt.Printf("   Route: %s (%s) -> %s (%s)\n",
			flight.Departure.City, flight.Departure.Airport,
			flight.Arrival.City, flight.Arrival.Airport)
		fmt.Printf("   Departure: %s\n", flight.Departure.Datetime)
		fmt.Printf("   Arrival: %s\n", flight.Arrival.Datetime)
		fmt.Printf("   Duration: %s (%d mins)\n", flight.Duration.Formatted, flight.Duration.TotalMinutes)
		fmt.Printf("   Stops: %d\n", flight.Stops)
		fmt.Printf("   Price: Rp%s\n", formatPrice(flight.Price.Amount))
		fmt.Printf("   Available Seats: %d\n", flight.AvailableSeats)
		fmt.Printf("   Provider: %s\n", flight.Provider)
	}

	fmt.Println()

	// Print full JSON of first flight for reference
	if len(response.Flights) > 0 {
		fmt.Println("Sample Flight JSON (first result):")
		fmt.Println("----------------------------------")
		jsonData, _ := json.MarshalIndent(response.Flights[0], "", "  ")
		fmt.Println(string(jsonData))
	}
}

// formatPrice formats price with thousands separator
func formatPrice(amount int) string {
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
