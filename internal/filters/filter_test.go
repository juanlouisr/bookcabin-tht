package filters

import (
	"testing"
	"time"

	"bookcabin/flight/internal/models"
)

func createTestFlights() []models.UnifiedFlight {
	return []models.UnifiedFlight{
		{
			ID:        "flight1",
			Price:     models.PriceInfo{Amount: 1000000},
			Duration:  models.DurationInfo{TotalMinutes: 120},
			Stops:     0,
			Departure: models.DepartureInfo{Timestamp: time.Date(2025, 12, 15, 8, 0, 0, 0, time.UTC).Unix()},
			Arrival:   models.ArrivalInfo{Timestamp: time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC).Unix()},
			Airline:   models.AirlineInfo{Name: "Garuda Indonesia"},
		},
		{
			ID:        "flight2",
			Price:     models.PriceInfo{Amount: 800000},
			Duration:  models.DurationInfo{TotalMinutes: 180},
			Stops:     1,
			Departure: models.DepartureInfo{Timestamp: time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC).Unix()},
			Arrival:   models.ArrivalInfo{Timestamp: time.Date(2025, 12, 15, 13, 0, 0, 0, time.UTC).Unix()},
			Airline:   models.AirlineInfo{Name: "Lion Air"},
		},
		{
			ID:        "flight3",
			Price:     models.PriceInfo{Amount: 1500000},
			Duration:  models.DurationInfo{TotalMinutes: 90},
			Stops:     0,
			Departure: models.DepartureInfo{Timestamp: time.Date(2025, 12, 15, 14, 0, 0, 0, time.UTC).Unix()},
			Arrival:   models.ArrivalInfo{Timestamp: time.Date(2025, 12, 15, 15, 30, 0, 0, time.UTC).Unix()},
			Airline:   models.AirlineInfo{Name: "Batik Air"},
		},
	}
}

func TestApplyFilters_MaxPrice(t *testing.T) {
	flights := createTestFlights()
	maxPrice := 1000000
	filters := models.FilterOptions{
		MaxPrice: &maxPrice,
	}

	result := ApplyFilters(flights, filters)

	if len(result) != 2 {
		t.Errorf("Expected 2 flights, got %d", len(result))
	}

	for _, f := range result {
		if f.Price.Amount > maxPrice {
			t.Errorf("Flight %s price %d exceeds max %d", f.ID, f.Price.Amount, maxPrice)
		}
	}
}

func TestApplyFilters_MinPrice(t *testing.T) {
	flights := createTestFlights()
	minPrice := 1000000
	filters := models.FilterOptions{
		MinPrice: &minPrice,
	}

	result := ApplyFilters(flights, filters)

	if len(result) != 2 {
		t.Errorf("Expected 2 flights, got %d", len(result))
	}

	for _, f := range result {
		if f.Price.Amount < minPrice {
			t.Errorf("Flight %s price %d below min %d", f.ID, f.Price.Amount, minPrice)
		}
	}
}

func TestApplyFilters_MaxStops(t *testing.T) {
	flights := createTestFlights()
	maxStops := 0
	filters := models.FilterOptions{
		MaxStops: &maxStops,
	}

	result := ApplyFilters(flights, filters)

	if len(result) != 2 {
		t.Errorf("Expected 2 flights (non-stop), got %d", len(result))
	}

	for _, f := range result {
		if f.Stops > maxStops {
			t.Errorf("Flight %s has %d stops, exceeds max %d", f.ID, f.Stops, maxStops)
		}
	}
}

func TestApplyFilters_Airlines(t *testing.T) {
	flights := createTestFlights()
	filters := models.FilterOptions{
		Airlines: []string{"Garuda Indonesia"},
	}

	result := ApplyFilters(flights, filters)

	if len(result) != 1 {
		t.Errorf("Expected 1 flight, got %d", len(result))
	}

	if result[0].Airline.Name != "Garuda Indonesia" {
		t.Errorf("Expected Garuda Indonesia, got %s", result[0].Airline.Name)
	}
}

func TestApplyFilters_MaxDuration(t *testing.T) {
	flights := createTestFlights()
	maxDuration := 120
	filters := models.FilterOptions{
		MaxDurationMins: &maxDuration,
	}

	result := ApplyFilters(flights, filters)

	if len(result) != 2 {
		t.Errorf("Expected 2 flights, got %d", len(result))
	}

	for _, f := range result {
		if f.Duration.TotalMinutes > maxDuration {
			t.Errorf("Flight %s duration %d exceeds max %d", f.ID, f.Duration.TotalMinutes, maxDuration)
		}
	}
}

func TestSortFlights_ByPriceAsc(t *testing.T) {
	flights := createTestFlights()

	result := SortFlights(flights, models.SortByPriceAsc)

	if len(result) != 3 {
		t.Fatalf("Expected 3 flights, got %d", len(result))
	}

	if result[0].Price.Amount != 800000 {
		t.Errorf("Expected cheapest flight first (800000), got %d", result[0].Price.Amount)
	}
	if result[2].Price.Amount != 1500000 {
		t.Errorf("Expected most expensive flight last (1500000), got %d", result[2].Price.Amount)
	}
}

func TestSortFlights_ByPriceDesc(t *testing.T) {
	flights := createTestFlights()

	result := SortFlights(flights, models.SortByPriceDesc)

	if len(result) != 3 {
		t.Fatalf("Expected 3 flights, got %d", len(result))
	}

	if result[0].Price.Amount != 1500000 {
		t.Errorf("Expected most expensive flight first (1500000), got %d", result[0].Price.Amount)
	}
	if result[2].Price.Amount != 800000 {
		t.Errorf("Expected cheapest flight last (800000), got %d", result[2].Price.Amount)
	}
}

func TestSortFlights_ByDurationAsc(t *testing.T) {
	flights := createTestFlights()

	result := SortFlights(flights, models.SortByDurationAsc)

	if len(result) != 3 {
		t.Fatalf("Expected 3 flights, got %d", len(result))
	}

	if result[0].Duration.TotalMinutes != 90 {
		t.Errorf("Expected shortest flight first (90), got %d", result[0].Duration.TotalMinutes)
	}
	if result[2].Duration.TotalMinutes != 180 {
		t.Errorf("Expected longest flight last (180), got %d", result[2].Duration.TotalMinutes)
	}
}

func TestSortFlights_ByDepartureAsc(t *testing.T) {
	flights := createTestFlights()

	result := SortFlights(flights, models.SortByDepartureAsc)

	if len(result) != 3 {
		t.Fatalf("Expected 3 flights, got %d", len(result))
	}

	if result[0].ID != "flight1" {
		t.Errorf("Expected earliest departure flight first (flight1), got %s", result[0].ID)
	}
	if result[2].ID != "flight3" {
		t.Errorf("Expected latest departure flight last (flight3), got %s", result[2].ID)
	}
}

func TestApplyFilters_EmptyFilters(t *testing.T) {
	flights := createTestFlights()
	filters := models.FilterOptions{}

	result := ApplyFilters(flights, filters)

	if len(result) != len(flights) {
		t.Errorf("Expected all flights to pass with empty filters, got %d", len(result))
	}
}
