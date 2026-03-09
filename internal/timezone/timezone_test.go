package timezone

import (
	"testing"
	"time"
)

func TestGetTimezoneByAirport_WIB(t *testing.T) {
	wibAirports := []string{"CGK", "HLP", "BDO", "SUB", "JOG", "SOC", "SRG", "BPN", "PKU", "PDG", "BTH", "PLM", "TKG"}

	for _, airport := range wibAirports {
		tz := GetTimezoneByAirport(airport)
		if tz.Code != "WIB" {
			t.Errorf("Expected WIB for %s, got %s", airport, tz.Code)
		}
		if tz.OffsetHours != 7 {
			t.Errorf("Expected offset +7 for %s, got %d", airport, tz.OffsetHours)
		}
	}
}

func TestGetTimezoneByAirport_WITA(t *testing.T) {
	witaAirports := []string{"DPS", "LOP", "BWN"}

	for _, airport := range witaAirports {
		tz := GetTimezoneByAirport(airport)
		if tz.Code != "WITA" {
			t.Errorf("Expected WITA for %s, got %s", airport, tz.Code)
		}
		if tz.OffsetHours != 8 {
			t.Errorf("Expected offset +8 for %s, got %d", airport, tz.OffsetHours)
		}
	}
}

func TestGetTimezoneByAirport_WIT(t *testing.T) {
	witAirports := []string{"BIK", "DJJ", "SOQ", "MKW", "TIM"}

	for _, airport := range witAirports {
		tz := GetTimezoneByAirport(airport)
		if tz.Code != "WIT" {
			t.Errorf("Expected WIT for %s, got %s", airport, tz.Code)
		}
		if tz.OffsetHours != 9 {
			t.Errorf("Expected offset +9 for %s, got %d", airport, tz.OffsetHours)
		}
	}
}

func TestGetTimezoneByAirport_Unknown(t *testing.T) {
	// Unknown airports should default to WIB
	tz := GetTimezoneByAirport("UNKNOWN")
	if tz.Code != "WIB" {
		t.Errorf("Expected WIB for unknown airport, got %s", tz.Code)
	}
}

func TestGetTimezoneByCity_WIB(t *testing.T) {
	wibCities := []string{"Jakarta", "Bandung", "Surabaya", "Yogyakarta", "Solo", "Semarang", "Palembang", "Medan", "Padang", "Batam"}

	for _, city := range wibCities {
		tz := GetTimezoneByCity(city)
		if tz.Code != "WIB" {
			t.Errorf("Expected WIB for %s, got %s", city, tz.Code)
		}
	}
}

func TestGetTimezoneByCity_WITA(t *testing.T) {
	witaCities := []string{"Denpasar", "Makassar", "Manado"}

	for _, city := range witaCities {
		tz := GetTimezoneByCity(city)
		if tz.Code != "WITA" {
			t.Errorf("Expected WITA for %s, got %s", city, tz.Code)
		}
	}
}

func TestGetTimezoneByCity_WIT(t *testing.T) {
	witCities := []string{"Jayapura", "Sorong"}

	for _, city := range witCities {
		tz := GetTimezoneByCity(city)
		if tz.Code != "WIT" {
			t.Errorf("Expected WIT for %s, got %s", city, tz.Code)
		}
	}
}

func TestGetTimezoneByCity_Unknown(t *testing.T) {
	// Unknown cities should default to WIB
	tz := GetTimezoneByCity("UnknownCity")
	if tz.Code != "WIB" {
		t.Errorf("Expected WIB for unknown city, got %s", tz.Code)
	}
}

func TestConvertToTimezone(t *testing.T) {
	// Create a time in UTC
	utcTime := time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC)

	// Convert to WIB (UTC+7)
	wibTime := ConvertToTimezone(utcTime, WIBInfo)
	if wibTime.Hour() != 19 {
		t.Errorf("Expected 19:00 WIB, got %d:00", wibTime.Hour())
	}

	// Convert to WITA (UTC+8)
	witaTime := ConvertToTimezone(utcTime, WITAInfo)
	if witaTime.Hour() != 20 {
		t.Errorf("Expected 20:00 WITA, got %d:00", witaTime.Hour())
	}

	// Convert to WIT (UTC+9)
	witTime := ConvertToTimezone(utcTime, WITInfo)
	if witTime.Hour() != 21 {
		t.Errorf("Expected 21:00 WIT, got %d:00", witTime.Hour())
	}
}

func TestFormatInTimezone(t *testing.T) {
	utcTime := time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC)

	formatted := FormatInTimezone(utcTime, WIBInfo, "2006-01-02 15:04")
	expected := "2025-12-15 19:00"
	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}
}

func TestTimezoneInfo_Location(t *testing.T) {
	// Test that locations are properly initialized
	if WIBInfo.Location == nil {
		t.Error("WIBInfo.Location should not be nil")
	}
	if WITAInfo.Location == nil {
		t.Error("WITAInfo.Location should not be nil")
	}
	if WITInfo.Location == nil {
		t.Error("WITInfo.Location should not be nil")
	}
}

func TestTimezoneOffsets(t *testing.T) {
	// WIB: UTC+7
	wibOffset := 7 * 60 * 60 // 7 hours in seconds
	if WIBInfo.Location.String() != WIB {
		// Check offset manually
		now := time.Now().In(WIBInfo.Location)
		_, offset := now.Zone()
		if offset != wibOffset {
			t.Errorf("Expected WIB offset %d, got %d", wibOffset, offset)
		}
	}

	// WITA: UTC+8
	witaOffset := 8 * 60 * 60 // 8 hours in seconds
	if WITAInfo.Location.String() != WITA {
		now := time.Now().In(WITAInfo.Location)
		_, offset := now.Zone()
		if offset != witaOffset {
			t.Errorf("Expected WITA offset %d, got %d", witaOffset, offset)
		}
	}

	// WIT: UTC+9
	witOffset := 9 * 60 * 60 // 9 hours in seconds
	if WITInfo.Location.String() != WIT {
		now := time.Now().In(WITInfo.Location)
		_, offset := now.Zone()
		if offset != witOffset {
			t.Errorf("Expected WIT offset %d, got %d", witOffset, offset)
		}
	}
}

func TestConvertFlightTime(t *testing.T) {
	// Test common flight route: CGK (WIB) to DPS (WITA)
	departureTime := time.Date(2025, 12, 15, 8, 0, 0, 0, WIBInfo.Location)
	arrivalTime := time.Date(2025, 12, 15, 11, 0, 0, 0, WITAInfo.Location)

	// Convert both to same timezone for duration calculation
	departureInWITA := ConvertToTimezone(departureTime, WITAInfo)
	arrivalInWITA := ConvertToTimezone(arrivalTime, WITAInfo)

	duration := arrivalInWITA.Sub(departureInWITA)
	expectedDuration := 2 * time.Hour

	if duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, duration)
	}
}
