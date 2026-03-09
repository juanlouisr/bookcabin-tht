# Flight Search & Aggregation System

A Go-based flight search and aggregation system that combines flight data from multiple airline APIs, processes and filters results, and returns optimized search results.

## Features

- **Multi-Provider Aggregation**: Aggregates flight data from Garuda Indonesia, Lion Air, Batik Air, and AirAsia
- **Protocol Agnostic**: Supports HTTP, SOAP, gRPC, or any transport protocol via the Transport interface
- **Provider-Specific Parsing**: Each provider handles its own data format conversion
- **Search & Filter**: Search by origin, destination, date with filters for price, stops, duration, airlines, and time ranges
- **Sorting Options**: Sort by price, duration, departure time, arrival time, or best value score
- **Best Value Algorithm**: Intelligent scoring based on price, duration, and number of stops
- **Caching**: In-memory caching with TTL for improved performance
- **Concurrent Provider Queries**: Parallel fetching from all providers with timeout handling
- **Error Handling**: Graceful handling of provider failures (e.g., AirAsia's 90% success rate simulation)
- **Rate Limiting**: Token bucket rate limiter for provider API protection
- **Retry Logic**: Exponential backoff with jitter for failed requests
- **Timezone Support**: WIB (UTC+7), WITA (UTC+8), WIT (UTC+9) conversions
- **Multi-City Search**: Support for complex multi-leg journeys
- **Round-Trip Search**: Support for return date searches

## Project Structure

```
.
├── internal/
│   ├── models/         # Domain models and types
│   │   ├── flight.go   # Flight, SearchRequest, SearchResponse models
│   │   └── provider.go # Provider interface
│   ├── transport/      # Transport layer (protocol abstraction)
│   │   ├── transport.go   # Transport interface (returns raw bytes)
│   │   ├── http.go        # HTTP transport implementation
│   │   ├── enhanced_http.go # HTTP transport with rate limiting & retry
│   │   └── mock.go        # Mock transport for testing
│   ├── providers/      # Provider implementations
│   │   ├── garuda.go   # Garuda Indonesia provider + parser
│   │   ├── lion.go     # Lion Air provider + parser
│   │   ├── batik.go    # Batik Air provider + parser
│   │   ├── airasia.go  # AirAsia provider + parser
│   │   ├── factory.go  # Provider factory with DI
│   │   └── util.go     # Shared provider utilities
│   ├── aggregator/     # Aggregation service
│   │   └── service.go  # Main aggregation logic
│   ├── filters/        # Filter and sort logic
│   │   └── filter.go   # Filter and sort implementations
│   ├── cache/          # Caching layer
│   │   └── cache.go    # In-memory cache with TTL
│   ├── handlers/       # HTTP handlers
│   │   └── search.go   # Search API endpoints
│   ├── ratelimit/      # Rate limiting
│   │   └── ratelimit.go # Token bucket implementation
│   ├── retry/          # Retry logic
│   │   └── retry.go    # Exponential backoff with jitter
│   └── timezone/       # Timezone handling
│       └── timezone.go # WIB, WITA, WIT support
├── main.go             # Application entry point
├── go.mod              # Go module definition
└── README.md           # This file
```

## Architecture

### Transport-Provider Separation

The system separates **transport** (how data is fetched) from **parsing** (how data is interpreted):

```
┌─────────────────────────────────────────────────────────────┐
│                     Provider Layer                          │
│  Each provider encapsulates its own parsing logic           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Garuda    │  │  Lion Air   │  │  AirAsia    │         │
│  │  - parse()  │  │  - parse()  │  │  - parse()  │         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │
└─────────┼────────────────┼────────────────┼────────────────┘
           │                │                │
           ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                     Transport Layer                         │
│    Returns raw bytes - protocol agnostic                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   HTTP      │  │   Mock      │  │    SOAP     │         │
│  │ Transport   │  │ Transport   │  │  Transport  │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### Why This Architecture?

1. **Provider Knows Its Format**: Each airline has unique data formats. The provider is the best place to parse its own format.
2. **Transport Agnostic**: The transport layer only fetches raw bytes. It doesn't care about the data format.
3. **Protocol Flexibility**: Easy to support SOAP, gRPC, or other protocols by implementing the Transport interface.
4. **Testability**: Mock transport for testing without hitting real APIs.

## Installation & Setup

1. Clone or navigate to the project directory:

```bash
cd bookcabin/flight
```

2. Install dependencies:

```bash
go mod tidy
```

3. Run the application:

```bash
go run main.go
```

The server will start on port 8080 by default (configurable via `PORT` environment variable).

## API Endpoints

### Health Check

```bash
GET /health
```

### Search Flights

```bash
POST /api/v1/search
```

Request body:

```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departure_date": "2025-12-15",
  "return_date": null,
  "passengers": 1,
  "cabin_class": "economy",
  "filters": {
    "max_price": 1500000,
    "max_stops": 1,
    "airlines": ["Garuda Indonesia", "Lion Air"]
  },
  "sort_by": "price_asc"
}
```

### Multi-City Search

```bash
POST /api/v1/search/multi-city
```

Request body:

```json
{
  "legs": [
    {
      "origin": "CGK",
      "destination": "DPS",
      "departure_date": "2025-12-15"
    },
    {
      "origin": "DPS",
      "destination": "SUB",
      "departure_date": "2025-12-20"
    },
    {
      "origin": "SUB",
      "destination": "CGK",
      "departure_date": "2025-12-25"
    }
  ],
  "passengers": 1,
  "cabin_class": "economy"
}
```

#### Filter Options

| Parameter           | Type     | Description                        |
| ------------------- | -------- | ---------------------------------- |
| `max_price`         | int      | Maximum price in IDR               |
| `min_price`         | int      | Minimum price in IDR               |
| `max_stops`         | int      | Maximum number of stops            |
| `airlines`          | []string | Filter by airline names or codes   |
| `departure_after`   | string   | Departure after time (HH:MM)       |
| `departure_before`  | string   | Departure before time (HH:MM)      |
| `arrival_after`     | string   | Arrival after time (HH:MM)         |
| `arrival_before`    | string   | Arrival before time (HH:MM)        |
| `max_duration_mins` | int      | Maximum flight duration in minutes |

#### Sort Options

| Option           | Description                            |
| ---------------- | -------------------------------------- |
| `price_asc`      | Price: lowest to highest (default)     |
| `price_desc`     | Price: highest to lowest               |
| `duration_asc`   | Duration: shortest to longest          |
| `duration_desc`  | Duration: longest to shortest          |
| `departure_asc`  | Departure time: earliest to latest     |
| `departure_desc` | Departure time: latest to earliest     |
| `arrival_asc`    | Arrival time: earliest to latest       |
| `arrival_desc`   | Arrival time: latest to earliest       |
| `best_value`     | Best value score (price + convenience) |

## Example Usage

### Basic Search

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15"
  }'
```

### Search with Filters

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "filters": {
      "max_price": 1000000,
      "max_stops": 0
    },
    "sort_by": "best_value"
  }'
```

### Multi-City Search

```bash
curl -X POST http://localhost:8080/api/v1/search/multi-city \
  -H "Content-Type: application/json" \
  -d '{
    "legs": [
      {"origin": "CGK", "destination": "DPS", "departure_date": "2025-12-15"},
      {"origin": "DPS", "destination": "SUB", "departure_date": "2025-12-20"}
    ],
    "passengers": 1,
    "cabin_class": "economy"
  }'
```

### Round-Trip Search

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "return_date": "2025-12-20",
    "passengers": 1,
    "cabin_class": "economy"
  }'
```

## Response Format

```json
{
  "search_criteria": {
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy"
  },
  "metadata": {
    "total_results": 10,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 285,
    "cache_hit": false
  },
  "flights": [
    {
      "id": "QZ7250_AirAsia",
      "provider": "AirAsia",
      "airline": {
        "name": "AirAsia",
        "code": "QZ"
      },
      "flight_number": "QZ7250",
      "departure": {
        "airport": "CGK",
        "city": "Jakarta",
        "datetime": "2025-12-15T15:15:00+07:00",
        "timestamp": 1734246900
      },
      "arrival": {
        "airport": "DPS",
        "city": "Denpasar",
        "datetime": "2025-12-15T20:35:00+08:00",
        "timestamp": 1734267300
      },
      "duration": {
        "total_minutes": 260,
        "formatted": "4h 20m"
      },
      "stops": 1,
      "price": {
        "amount": 485000,
        "currency": "IDR"
      },
      "available_seats": 88,
      "cabin_class": "economy",
      "aircraft": null,
      "amenities": [],
      "baggage": {
        "carry_on": "Cabin baggage only",
        "checked": "Additional fee"
      },
      "best_value_score": 0.75
    }
  ]
}
```

## Bonus Features

### Rate Limiting

The system includes a token bucket rate limiter (`internal/ratelimit/ratelimit.go`) to protect provider APIs from being overwhelmed:

```go
// Create a rate limiter with 10 requests per second, burst of 20
limiter := ratelimit.NewTokenBucket(10.0, 20)

// Wait for a token before making a request
err := limiter.Wait(ctx)
```

### Retry Logic with Exponential Backoff

Failed requests are automatically retried with exponential backoff and jitter (`internal/retry/retry.go`):

```go
config := retry.Config{
    MaxRetries:          3,
    InitialInterval:     500 * time.Millisecond,
    MaxInterval:         5 * time.Second,
    Multiplier:          2.0,
    RandomizationFactor: 0.1,
}

err := retry.Do(ctx, retryableFunc, isRetryableFunc, config)
```

### Timezone Support

Indonesian timezone handling for WIB, WITA, and WIT (`internal/timezone/timezone.go`):

```go
// Get timezone by airport code
tz := timezone.GetTimezoneByAirport("CGK") // Returns WIB (UTC+7)
tz := timezone.GetTimezoneByAirport("DPS") // Returns WITA (UTC+8)
tz := timezone.GetTimezoneByAirport("DJJ") // Returns WIT (UTC+9)

// Convert time to specific timezone
localTime := timezone.ConvertToTimezone(utcTime, tz)
```

### Enhanced HTTP Transport

The `EnhancedHTTPTransport` combines rate limiting and retry logic:

```go
transport := transport.NewEnhancedHTTPTransport(
    "https://api.airline.com",
    apiKey,
    10*time.Second,  // timeout
    10.0,            // rate limit (requests per second)
    20,              // burst
)
```

## Switching to Real APIs

### 1. Set Environment Variable

```bash
export USE_REAL_PROVIDERS=true
```

### 2. Configure Real HTTP Transports

Update `internal/providers/factory.go`:

```go
func createRealProviders() ([]models.Provider, error) {
    var providers []models.Provider

    // Garuda Indonesia
    garudaTransport := transport.NewHTTPTransport(
        "https://api.garuda-indonesia.com/v1",
        os.Getenv("GARUDA_API_KEY"),
        10*time.Second,
    )
    garuda := NewGarudaProvider(garudaTransport)
    providers = append(providers, garuda)

    // Lion Air
    lionTransport := transport.NewHTTPTransport(
        "https://api.lionair.co.id/v1",
        os.Getenv("LION_AIR_API_KEY"),
        10*time.Second,
    )
    lionAir := NewLionAirProvider(lionTransport)
    providers = append(providers, lionAir)

    // Batik Air
    batikTransport := transport.NewHTTPTransport(
        "https://api.batikair.com/v1",
        os.Getenv("BATIK_AIR_API_KEY"),
        10*time.Second,
    )
    batikAir := NewBatikAirProvider(batikTransport)
    providers = append(providers, batikAir)

    // AirAsia
    airAsiaTransport := transport.NewHTTPTransport(
        "https://api.airasia.com/v1",
        os.Getenv("AIRASIA_API_KEY"),
        10*time.Second,
    )
    airAsia := NewAirAsiaProvider(airAsiaTransport)
    providers = append(providers, airAsia)

    return providers, nil
}
```

### Using Enhanced Transport with Rate Limiting and Retry

```go
// Create enhanced transport with rate limiting and retry
enhancedTransport := transport.NewEnhancedHTTPTransport(
    "https://api.garuda-indonesia.com/v1",
    os.Getenv("GARUDA_API_KEY"),
    10*time.Second,
    10.0, // 10 requests per second
    20,   // burst of 20
)
garuda := NewGarudaProvider(enhancedTransport)
```

### Adding a New Transport Protocol (e.g., SOAP)

1. Create a new transport implementation:

```go
// internal/transport/soap.go
type SOAPTransport struct {
    Endpoint string
    Client   *soap.Client
}

func (t *SOAPTransport) Fetch(ctx context.Context, req Request) (*Response, error) {
    // Make SOAP call
    // Return raw bytes
}
```

2. Use it in the factory:

```go
soapTransport := &transport.SOAPTransport{
    Endpoint: "https://api.airline.com/soap",
}
provider := NewAirlineProvider(soapTransport)
```

The provider doesn't need to change - it just receives raw bytes and parses them.

## Design Decisions

### 1. Transport Returns Raw Bytes

The transport layer should not parse data. It only fetches raw bytes and returns them. This allows:

- Supporting any protocol (HTTP, SOAP, gRPC, etc.)
- Easy testing with mock transports
- Provider independence from transport details

### 2. Provider Handles Parsing

Each provider knows its own data format and is responsible for parsing. This allows:

- Provider-specific JSON/XML schemas
- Different API versions per provider
- Easy addition of new providers without changing transport code

### 3. Transport Interface

```go
type Transport interface {
    Fetch(ctx context.Context, request Request) (*Response, error)
    GetTimeout() time.Duration
}
```

This simple interface can be implemented for any protocol.

### 4. Best Value Algorithm

The best value score is calculated using normalized values:

- 50% weight on price
- 30% weight on duration
- 20% weight on number of stops

Lower scores indicate better value.

### 5. Caching Strategy

Simple in-memory cache with TTL for search results. Cache key is generated from search criteria (origin, destination, date, cabin class).

### 6. Error Handling

Provider failures are tracked but don't fail the entire search. Failed providers are reported in metadata, allowing the system to return partial results.

### 7. IDR Price Formatting

Prices are formatted with Indonesian thousands separator (e.g., "Rp 1.234.567"):

```go
formatted := models.FormatPrice(1234567) // Returns "1.234.567"
```

## Provider Characteristics

| Provider         | Delay     | Success Rate | Notes                               |
| ---------------- | --------- | ------------ | ----------------------------------- |
| Garuda Indonesia | 50-100ms  | 100%         | Premium airline with full amenities |
| Lion Air         | 100-200ms | 100%         | Budget airline, basic services      |
| Batik Air        | 200-400ms | 100%         | Full service, snack/meal included   |
| AirAsia          | 50-150ms  | 90%          | Budget airline, occasional failures |

## Future Enhancements

- [x] Round-trip search support
- [x] Multi-city search support
- [x] Advanced timezone handling (WIB, WITA, WIT)
- [x] Rate limiting for provider APIs
- [x] Retry logic with exponential backoff
- [ ] Database persistence for search history
- [ ] Advanced caching (Redis)
- [ ] Real-time price updates via WebSockets
- [ ] SOAP transport implementation
- [ ] gRPC transport implementation
