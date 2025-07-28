package util

import "spacetraders/constant"

func GetFuelCost(distance int, flightMode constant.FlightMode) int {
	return flightMode.GetFuelCost(distance)
}

func GetFlightTime(distance int, flightMode constant.FlightMode) float64 {
	return flightMode.GetFlightTime(distance)
}
