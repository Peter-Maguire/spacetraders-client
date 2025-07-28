package constant

type FlightMode string

const (
	FlightModeCruise  FlightMode = "CRUISE"
	FlightModeDrift   FlightMode = "DRIFT"
	FlightModeStealth FlightMode = "STEALTH"
	FlightModeBurn    FlightMode = "BURN"
)

func (fm FlightMode) GetFuelCost(distance int) int {
	switch fm {
	case FlightModeCruise, FlightModeStealth:
		return distance
	case FlightModeDrift:
		return 1
	case FlightModeBurn:
		return 2 * distance
	}
	return 9999
}

func (fm FlightMode) GetFlightTime(distance int) float64 {
	d := float64(distance)
	switch fm {
	case FlightModeCruise:
		return 15 + (0.333 * d)
	case FlightModeDrift:
		return 15 + (3.333 * d)
	case FlightModeBurn:
		return 15 + (0.167 * d)
	case FlightModeStealth:
		return 15 + (0.667 * d)
	}
	return 9999
}
