package routine

import (
	"fmt"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/util"
)

type Rescue struct {
	shipSymbol string
}

func (r Rescue) Run(state *State) RoutineResult {

	var ship *entity.Ship
	for _, st := range *state.States {
		if st.Ship.Symbol == r.shipSymbol {
			ship = st.Ship
			break
		}
	}
	if ship == nil {
		state.Log("Couldn't find ship that I'm supposed to be rescuing")
		return RoutineResult{
			Stop:       true,
			StopReason: "Unable to find ship to rescue",
		}
	}

	state.Ship.GetCargo(state.Context)

	fuelCargo := state.Ship.Cargo.GetSlotWithItem("FUEL")
	if fuelCargo == nil {
		wp := database.GetWaypoint(state.Ship.Nav.WaypointSymbol)
		wpData := wp.GetData()

		market := util.GetClosestMarketSelling([]string{"FUEL"}, wpData.LimitedWaypointData)

		if market.GetSystemName() != state.Ship.Nav.SystemSymbol {
			return RoutineResult{
				Stop:       true,
				StopReason: "You done messed it up again",
			}
		}

		if market == nil {
			return RoutineResult{
				Stop:       true,
				StopReason: "Unable to find market selling fuel",
			}
		}

		if state.Ship.Nav.WaypointSymbol != *market {
			state.Log("Going to market that sells Fuel")
			return RoutineResult{
				SetRoutine: NavigateTo{waypoint: *market, next: r},
			}
		}

		marketData, _ := market.GetMarket(state.Context)
		database.UpdateMarketRates(*market, marketData.TradeGoods)

		fuelGood := marketData.GetTradeGood("FUEL")
		if fuelGood == nil {
			state.Log("Fuel isn't being sold here")
			return RoutineResult{}
		}

		_, err := state.Ship.Purchase(state.Context, "FUEL", 1)
		if err != nil {
			state.Log(err.Message)
			return RoutineResult{
				Stop:       true,
				StopReason: "Unable to purchase fuel " + err.Message,
			}
		}

		state.Ship.Refuel(state.Context)
	}

	if state.Ship.Nav.WaypointSymbol != ship.Nav.WaypointSymbol {
		return RoutineResult{
			SetRoutine: NavigateTo{waypoint: ship.Nav.WaypointSymbol, next: r},
		}
	}

	state.Ship.EnsureNavState(state.Context, "ORBIT")
	err := state.Ship.TransferCargo(state.Context, ship.Symbol, "FUEL", 1)
	if err != nil {
		state.Log("Transfer cargo failed: " + err.Message)
	}

	return RoutineResult{
		SetRoutine: DetermineObjective{},
	}
}

func (r Rescue) Name() string {

	return fmt.Sprintf("Rescue %s", r.shipSymbol)
}
