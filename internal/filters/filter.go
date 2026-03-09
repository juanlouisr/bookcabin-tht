package filters

import (
	"sort"
	"strings"
	"time"

	"bookcabin/flight/internal/models"
)

// ApplyFilters applies filter criteria to flights
func ApplyFilters(flights []models.UnifiedFlight, filters models.FilterOptions) []models.UnifiedFlight {
	if len(flights) == 0 {
		return flights
	}

	var result []models.UnifiedFlight

	for _, flight := range flights {
		if matchesFilter(flight, filters) {
			result = append(result, flight)
		}
	}

	return result
}

// matchesFilter checks if a flight matches the filter criteria
func matchesFilter(flight models.UnifiedFlight, filters models.FilterOptions) bool {
	// Price filters
	if filters.MinPrice != nil && flight.Price.Amount < *filters.MinPrice {
		return false
	}
	if filters.MaxPrice != nil && flight.Price.Amount > *filters.MaxPrice {
		return false
	}

	// Stops filter
	if filters.MaxStops != nil && flight.Stops > *filters.MaxStops {
		return false
	}

	// Airlines filter
	if len(filters.Airlines) > 0 {
		found := false
		for _, airline := range filters.Airlines {
			if strings.EqualFold(flight.Airline.Name, airline) ||
				strings.EqualFold(flight.Airline.Code, airline) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Departure time filters
	depTime := time.Unix(flight.Departure.Timestamp, 0)
	if filters.DepartureAfter != nil && depTime.Before(*filters.DepartureAfter) {
		return false
	}
	if filters.DepartureBefore != nil && depTime.After(*filters.DepartureBefore) {
		return false
	}

	// Arrival time filters
	arrTime := time.Unix(flight.Arrival.Timestamp, 0)
	if filters.ArrivalAfter != nil && arrTime.Before(*filters.ArrivalAfter) {
		return false
	}
	if filters.ArrivalBefore != nil && arrTime.After(*filters.ArrivalBefore) {
		return false
	}

	// Duration filter
	if filters.MaxDurationMins != nil && flight.Duration.TotalMinutes > *filters.MaxDurationMins {
		return false
	}

	return true
}

// SortFlights sorts flights based on the specified option
func SortFlights(flights []models.UnifiedFlight, sortBy models.SortOption) []models.UnifiedFlight {
	switch sortBy {
	case models.SortByPriceAsc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount < flights[j].Price.Amount
		})
	case models.SortByPriceDesc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount > flights[j].Price.Amount
		})
	case models.SortByDurationAsc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Duration.TotalMinutes < flights[j].Duration.TotalMinutes
		})
	case models.SortByDurationDesc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Duration.TotalMinutes > flights[j].Duration.TotalMinutes
		})
	case models.SortByDepartureAsc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Departure.Timestamp < flights[j].Departure.Timestamp
		})
	case models.SortByDepartureDesc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Departure.Timestamp > flights[j].Departure.Timestamp
		})
	case models.SortByArrivalAsc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Arrival.Timestamp < flights[j].Arrival.Timestamp
		})
	case models.SortByArrivalDesc:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Arrival.Timestamp > flights[j].Arrival.Timestamp
		})
	case models.SortByBestValue:
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].BestValueScore < flights[j].BestValueScore
		})
	}

	return flights
}
