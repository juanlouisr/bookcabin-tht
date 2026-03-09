package timezone

import (
	"fmt"
	"time"
)

// Indonesia Timezone Constants
const (
	WIB  = "Asia/Jakarta"  // Western Indonesian Time (UTC+7) - Jakarta, Sumatra, Java, West/Central Kalimantan
	WITA = "Asia/Makassar" // Central Indonesian Time (UTC+8) - Bali, Nusa Tenggara, South/East/North Kalimantan, Sulawesi
	WIT  = "Asia/Jayapura" // Eastern Indonesian Time (UTC+9) - Maluku, Papua
)

// TimezoneInfo holds information about an Indonesian timezone
type TimezoneInfo struct {
	Code        string
	Name        string
	Location    *time.Location
	OffsetHours int
}

var (
	// WIBInfo Western Indonesian Time
	WIBInfo = &TimezoneInfo{
		Code:        "WIB",
		Name:        "Western Indonesian Time",
		OffsetHours: 7,
	}

	// WITAInfo Central Indonesian Time
	WITAInfo = &TimezoneInfo{
		Code:        "WITA",
		Name:        "Central Indonesian Time",
		OffsetHours: 8,
	}

	// WITInfo Eastern Indonesian Time
	WITInfo = &TimezoneInfo{
		Code:        "WIT",
		Name:        "Eastern Indonesian Time",
		OffsetHours: 9,
	}
)

func init() {
	var err error
	WIBInfo.Location, err = time.LoadLocation(WIB)
	if err != nil {
		// Fallback if timezone database not available
		WIBInfo.Location = time.FixedZone("WIB", 7*60*60)
	}

	WITAInfo.Location, err = time.LoadLocation(WITA)
	if err != nil {
		WITAInfo.Location = time.FixedZone("WITA", 8*60*60)
	}

	WITInfo.Location, err = time.LoadLocation(WIT)
	if err != nil {
		WITInfo.Location = time.FixedZone("WIT", 9*60*60)
	}
}

// ConvertToTimezone converts a time to the specified Indonesian timezone
func ConvertToTimezone(t time.Time, tz *TimezoneInfo) time.Time {
	return t.In(tz.Location)
}

// FormatInTimezone formats a time in the specified Indonesian timezone
func FormatInTimezone(t time.Time, tz *TimezoneInfo, layout string) string {
	return t.In(tz.Location).Format(layout)
}

// GetTimezoneByAirport returns the timezone for a given Indonesian airport code
func GetTimezoneByAirport(airportCode string) *TimezoneInfo {
	switch airportCode {
	// WIB (UTC+7) - Java, Sumatra, West/Central Kalimantan
	case "CGK", "HLP", "BDO", "SUB", "JOG", "SOC", "SRG", "BPN", "PKU", "PDG", "BTH", "PLM", "TKG":
		return WIBInfo

	// WITA (UTC+8) - Bali, Nusa Tenggara, Sulawesi, East/South/North Kalimantan
	case "DPS", "LOP", "BWN":
		return WITAInfo

	// WIT (UTC+9) - Maluku, Papua
	case "BIK", "DJJ", "SOQ", "MKW", "TIM":
		return WITInfo

	default:
		// Default to WIB for unknown airports
		return WIBInfo
	}
}

// GetTimezoneByCity returns the timezone for a given Indonesian city
func GetTimezoneByCity(city string) *TimezoneInfo {
	switch city {
	// WIB Cities
	case "Jakarta", "Bandung", "Surabaya", "Yogyakarta", "Solo", "Semarang",
		"Medan", "Palembang", "Padang", "Pekanbaru", "Batam", "Bandar Lampung",
		"Pontianak", "Palangkaraya":
		return WIBInfo

	// WITA Cities
	case "Denpasar", "Mataram", "Kupang", "Makassar", "Manado", "Palu",
		"Kendari", "Gorontalo", "Samarinda", "Balikpapan", "Banjarmasin":
		return WITAInfo

	// WIT Cities
	case "Jayapura", "Manokwari", "Sorong", "Timika", "Merauke":
		return WITInfo

	default:
		return WIBInfo
	}
}

// FormatWithTimezoneCode formats time with timezone code (e.g., "15:04 WIB")
func FormatWithTimezoneCode(t time.Time, tz *TimezoneInfo) string {
	return fmt.Sprintf("%s %s", t.In(tz.Location).Format("15:04"), tz.Code)
}

// CalculateDurationWithTimezone calculates flight duration considering timezone differences
func CalculateDurationWithTimezone(departure time.Time, arrival time.Time, depTimezone, arrTimezone *TimezoneInfo) time.Duration {
	// Convert both times to UTC for accurate duration calculation
	depUTC := departure.In(time.UTC)
	arrUTC := arrival.In(time.UTC)

	return arrUTC.Sub(depUTC)
}

// FormatDuration formats duration to human readable string
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", minutes)
}

// IsSameTimezone checks if two airports are in the same timezone
func IsSameTimezone(airport1, airport2 string) bool {
	tz1 := GetTimezoneByAirport(airport1)
	tz2 := GetTimezoneByAirport(airport2)
	return tz1.OffsetHours == tz2.OffsetHours
}

// GetTimezoneDifference returns the hour difference between two timezones
func GetTimezoneDifference(tz1, tz2 *TimezoneInfo) int {
	return tz1.OffsetHours - tz2.OffsetHours
}
