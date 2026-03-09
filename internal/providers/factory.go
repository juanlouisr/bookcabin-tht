package providers

import (
	"os"

	"bookcabin/flight/internal/models"
	"bookcabin/flight/internal/transport"
)

// CreateProviders initializes all flight providers
// In production, set USE_REAL_PROVIDERS=true environment variable to use real HTTP transports
func CreateProviders() ([]models.Provider, error) {
	useRealProviders := os.Getenv("USE_REAL_PROVIDERS") == "true"

	if useRealProviders {
		return createRealProviders()
	}

	return createMockProviders()
}

// createMockProviders creates providers with mock transports
func createMockProviders() ([]models.Provider, error) {
	var providers []models.Provider

	// Create Garuda Indonesia provider with mock transport
	garudaTransport := transport.NewMockTransportFromFile(
		"spec/garuda_indonesia_search_response.json",
		50, 100, // delay range
		1.0, // success rate
	)
	garuda := NewGarudaProvider(garudaTransport)
	providers = append(providers, garuda)

	// Create Lion Air provider with mock transport
	lionTransport := transport.NewMockTransportFromFile(
		"spec/lion_air_search_response.json",
		100, 200, // delay range
		1.0, // success rate
	)
	lionAir := NewLionAirProvider(lionTransport)
	providers = append(providers, lionAir)

	// Create Batik Air provider with mock transport
	batikTransport := transport.NewMockTransportFromFile(
		"spec/batik_air_search_response.json",
		200, 400, // delay range
		1.0, // success rate
	)
	batikAir := NewBatikAirProvider(batikTransport)
	providers = append(providers, batikAir)

	// Create AirAsia provider with mock transport (90% success rate)
	airAsiaTransport := transport.NewMockTransportFromFile(
		"spec/airasia_search_response.json",
		50, 150, // delay range
		0.9, // success rate
	)
	airAsia := NewAirAsiaProvider(airAsiaTransport)
	providers = append(providers, airAsia)

	return providers, nil
}

// createRealProviders creates providers with real HTTP transports
// This is a template - update with actual API endpoints and credentials
func createRealProviders() ([]models.Provider, error) {
	// Example: Create Garuda Indonesia provider with real HTTP transport
	// Uncomment and configure when ready for production
	/*
		garudaTransport := transport.NewHTTPTransport(
			"https://api.garuda-indonesia.com/v1",
			os.Getenv("GARUDA_API_KEY"),
			10*time.Second,
		)
		garuda := NewGarudaProvider(garudaTransport)
		providers = append(providers, garuda)

		// Add other providers similarly...
		lionTransport := transport.NewHTTPTransport(
			"https://api.lionair.co.id/v1",
			os.Getenv("LION_AIR_API_KEY"),
			10*time.Second,
		)
		lionAir := NewLionAirProvider(lionTransport)
		providers = append(providers, lionAir)

		batikTransport := transport.NewHTTPTransport(
			"https://api.batikair.com/v1",
			os.Getenv("BATIK_AIR_API_KEY"),
			10*time.Second,
		)
		batikAir := NewBatikAirProvider(batikTransport)
		providers = append(providers, batikAir)

		airAsiaTransport := transport.NewHTTPTransport(
			"https://api.airasia.com/v1",
			os.Getenv("AIRASIA_API_KEY"),
			10*time.Second,
		)
		airAsia := NewAirAsiaProvider(airAsiaTransport)
		providers = append(providers, airAsia)
	*/

	// For now, fall back to mock providers
	return createMockProviders()
}
