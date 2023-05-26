package util

func GetFuelCost(distance int, flightMode string) int {
    switch flightMode {
    case "CRUISE", "STEALTH":
        return distance
    case "DRIFT":
        return 1
    case "BURN":
        return 2 * distance
    }
    return 9999
}

func GetFlightTime(distance int, flightMode string) float64 {
    d := float64(distance)
    switch flightMode {
    case "CRUISE":
        return 15 + (0.333 * d)
    case "DRIFT":
        return 15 + (3.333 * d)
    case "BURN":
        return 15 + (0.167 * d)
    case "STEALTH":
        return 15 + (0.667 * d)
    }
    return 9999
}
