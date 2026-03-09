package providers

// getCityFromAirport returns the city name for a given airport code
func getCityFromAirport(code string) string {
	switch code {
	case "CGK":
		return "Jakarta"
	case "DPS":
		return "Denpasar"
	case "SUB":
		return "Surabaya"
	case "UPG":
		return "Makassar"
	case "SOC":
		return "Solo"
	case "JOG":
		return "Yogyakarta"
	case "BDO":
		return "Bandung"
	case "PLM":
		return "Palembang"
	case "PDG":
		return "Padang"
	case "BTH":
		return "Batam"
	default:
		return code
	}
}
