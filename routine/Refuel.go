package routine

import (
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
)

type Refuel struct {
	next           Routine
	hasTriedMarket bool
}

func (r Refuel) Run(state *State) RoutineResult {

	if state.Ship.Fuel.IsFull() {
		// TODO: this implies we do this if we do have fuel already
		state.Log("We don't have the fuel to get to wherever we're going, so we drift there")
		_ = state.Ship.SetFlightMode(state.Context, "DRIFT")
		return RoutineResult{SetRoutine: r.next}
	}

	//state.Log(fmt.Sprintf("Has tried market: %v", r.hasTriedMarket))
	// TODO: rewrite this code for more efficient refuelling
	if !r.hasTriedMarket {
		state.Log("Seeing if we have a market here")
		market, err := state.Ship.Nav.WaypointSymbol.GetMarket(state.Context)
		if err == nil {
			go database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, market.TradeGoods)
		}
		if err != nil || market.GetTradeGood("FUEL") == nil {
			state.Log("No market here selling fuel")
			waypoints, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints(state.Context)
			for _, waypoint := range *waypoints {
				if waypoint.HasTrait("MARKETPLACE") && waypoint.Symbol != state.Ship.Nav.WaypointSymbol {
					state.Log("Trying a different market")
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
						SetRoutine: NavigateTo{
							waypoint: waypoint.Symbol,
							next:     Refuel{next: r.next},
						},
					}
				}
			}
		} else {
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
