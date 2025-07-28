package routine

import (
	"fmt"
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

	if state.Ship.Fuel.IsFull() {
		state.Log("Fuel is already full")
		_ = state.Ship.SetFlightMode(state.Context, "DRIFT")
		return RoutineResult{SetRoutine: r.next}
	}

	// TODO: rescue
	if state.Ship.Fuel.Current == 0 {

		rescueShips := make([]*State, 0)
		for _, st := range *state.States {
			if st.Ship.Cargo.Capacity == 0 ||
				st.Ship.Fuel.Current == 0 ||
				st.Ship.Nav.Status == "IN_TRANSIT" {
				continue
			}
			rescueShips = append(rescueShips, st)
		}

		if len(rescueShips) == 0 {
			state.Log("No ships are currently available to rescue me :'(")
			return RoutineResult{WaitSeconds: 120}
		}

		ourWaypoint := database.GetWaypoint(state.Ship.Nav.WaypointSymbol)
		ourWaypointData := ourWaypoint.GetData()

		sort.Slice(rescueShips, func(i, j int) bool {
			shipIWaypoint := database.GetWaypoint(rescueShips[i].Ship.Nav.WaypointSymbol)
			shipIWaypointData := shipIWaypoint.GetData()
			shipJWaypoint := database.GetWaypoint(rescueShips[j].Ship.Nav.WaypointSymbol)
			shipJWaypointData := shipJWaypoint.GetData()
			return shipIWaypointData.GetDistanceFrom(ourWaypointData.LimitedWaypointData) < shipJWaypointData.GetDistanceFrom(ourWaypointData.LimitedWaypointData)
		})

		closestShip := rescueShips[0]
		state.Log(fmt.Sprintf("Closest ship that can rescue us is %s", closestShip.Ship.Symbol))

		closestShip.ForceRoutine = Rescue{shipSymbol: state.Ship.Symbol}

		return RoutineResult{SetRoutine: AwaitRescue{next: r}}
	}

	// TODO: this would not be necessary if we properly handled refreshing the ship data
	ship, _ := state.Agent.GetShip(state.Context, state.Ship.Symbol)
	if ship != nil {
		state.Ship = ship
	}

	state.Log(fmt.Sprintf("Current fuel level: %d/%d", state.Ship.Fuel.Current, state.Ship.Fuel.Capacity))

	if !r.hasTriedMarket {
		state.Log("Seeing if we have a market here")
		market, err := state.Ship.Nav.WaypointSymbol.GetMarket(state.Context)
		if err == nil {
			go database.UpdateMarketRates(state.Ship.Nav.WaypointSymbol, market.TradeGoods)
		}
		if err != nil || market.GetTradeGood("FUEL") == nil {
			state.Log("No market or market doesn't sell fuel")
			marketsSellingFuel := database.GetMarketsSelling([]string{"FUEL"})

			currentWaypoint, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

			inRange := make([]entity.LimitedWaypointData, 0)
			for _, rate := range marketsSellingFuel {
				if rate.Waypoint.GetSystemName() != state.Ship.Nav.SystemSymbol {
					continue
				}
				rateWaypoint := entity.LimitedWaypointData{Symbol: rate.Waypoint, X: rate.WaypointX, Y: rate.WaypointY}
				distance := rateWaypoint.GetDistanceFrom(currentWaypoint.LimitedWaypointData)
				fuelCost := util.GetFuelCost(distance, state.Ship.Nav.FlightMode)
				if fuelCost <= state.Ship.Fuel.Current {
					inRange = append(inRange, rateWaypoint)
				}
			}

			if len(inRange) == 0 {
				if state.Ship.Nav.FlightMode == "DRIFT" {
					state.Log("We are already in drift mode")
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

			state.Log(fmt.Sprintf("Found market %s within range (%d)", inRange[0].Symbol, inRange[0].GetDistanceFrom(currentWaypoint.LimitedWaypointData)))

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
	return fmt.Sprintf("Refuel -> %s", r.next.Name())
}
