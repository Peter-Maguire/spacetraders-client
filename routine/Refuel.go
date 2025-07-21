package routine

import (
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/util"
)

type Refuel struct {
	next           Routine
	hasTriedMarket bool
}

func (r Refuel) Run(state *State) RoutineResult {

	//if state.Ship.Fuel.IsFull() {
	//	state.Log("Fuel is already full")
	//	_ = state.Ship.SetFlightMode(state.Context, "CRUISE")
	//	return RoutineResult{SetRoutine: r.next}
	//}

	ship, _ := state.Agent.GetShip(state.Context, state.Ship.Symbol)
	if ship != nil {
		state.Ship = ship
	}

	if !r.hasTriedMarket {
		state.Log("Seeing if we have a market here")
		market, err := state.Ship.Nav.WaypointSymbol.GetMarket(state.Context)
		if err == nil {
			go database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, market.TradeGoods)
		}
		if err != nil || market.GetTradeGood("FUEL") == nil {

			marketsSellingFuel := database.GetMarketsSelling([]string{"FUEL"})

			currentWaypoint, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

			inRange := make([]entity.LimitedWaypointData, 0)
			for _, rate := range marketsSellingFuel {
				rateWaypoint := entity.LimitedWaypointData{Symbol: rate.Waypoint, X: rate.WaypointX, Y: rate.WaypointY}
				distance := rateWaypoint.GetDistanceFrom(currentWaypoint.LimitedWaypointData)
				fuelCost := util.GetFuelCost(distance, state.Ship.Nav.FlightMode)
				if fuelCost <= state.Ship.Fuel.Current {
					inRange = append(inRange, rateWaypoint)
				}
			}

			if len(inRange) == 0 {
				if state.Ship.Nav.FlightMode == "DRIFT" {
					return RoutineResult{
						SetRoutine: FindNewWaypoint{
							desiredTrait: "MARKETPLACE",
							next:         r,
						},
					}
				}
				state.Log("Nowhere we can go without drifting")
				_ = state.Ship.SetFlightMode(state.Context, "DRIFT")
				return RoutineResult{}
			}

			sort.Slice(inRange, func(i, j int) bool {
				d1 := inRange[i].GetDistanceFrom(currentWaypoint.LimitedWaypointData)
				d2 := inRange[j].GetDistanceFrom(currentWaypoint.LimitedWaypointData)
				return d1 < d2
			})

			return RoutineResult{
				SetRoutine: NavigateTo{waypoint: inRange[0].Symbol, next: r},
			}
		}

		state.Log("Trying to refuel here")
		_ = state.Ship.EnsureNavState(state.Context, entity.NavDocked)
		refuelErr := state.Ship.Refuel(state.Context)

		if refuelErr == nil {
			if state.Ship.Nav.FlightMode == "DRIFT" {
				state.Log("Exiting drift mode")
				_ = state.Ship.SetFlightMode(state.Context, "CRUISE")
			}

			return RoutineResult{
				SetRoutine: r.next,
			}
		}

		switch refuelErr.Code {
		case http.ErrShipInTransit, http.ErrNavigateInTransit:
			state.Log("Ship in transit")
			return RoutineResult{
				WaitSeconds: 30,
			}
		}
		state.Log(refuelErr.Error())

	}

	state.Log("Cannot refuel")

	if state.Ship.Nav.FlightMode == "DRIFT" {
		state.Log("We're boned")
		return RoutineResult{
			Stop:       true,
			StopReason: "Unable to refuel",
		}
	}

	state.Log("Setting flight mode to drift")
	_ = state.Ship.SetFlightMode(state.Context, "DRIFT")

	return RoutineResult{
		SetRoutine: r.next,
	}

}

func (r Refuel) Name() string {
	return "Refuel"
}
