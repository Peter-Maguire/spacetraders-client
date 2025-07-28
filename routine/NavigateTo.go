package routine

import (
	"fmt"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/util"
	"time"
)

type NavigateTo struct {
	waypoint     entity.Waypoint
	next         Routine
	nextIfNoFuel Routine
}

func (n NavigateTo) Run(state *State) RoutineResult {
	state.Log(fmt.Sprint("Navigating to ", n.waypoint))

	if state.Ship.Nav.WaypointSymbol == n.waypoint && state.Ship.Nav.Status != "IN_TRANSIT" {
		state.Log("We're already at our destination")
		return RoutineResult{
			SetRoutine: n.next,
		}
	}

	if state.Ship.Nav.Status == "IN_TRANSIT" && state.Ship.Nav.Route.Arrival.After(time.Now()) {
		state.Log("We're on our way somewhere")
		return RoutineResult{
			WaitUntil: &state.Ship.Nav.Route.Arrival,
		}
	}

	if state.Ship.Fuel.Current == 1 {
		state.Log("Refuelling so we don't run out of fuel")
		return RoutineResult{
			SetRoutine: Refuel{next: n},
		}
	}

	if state.Ship.Nav.SystemSymbol != n.waypoint.GetSystemName() {
		fmt.Println(state.Ship.Nav.SystemSymbol, n.waypoint.GetSystemName())
		state.Log("Jumping to system first")
		return RoutineResult{
			SetRoutine: GoToSystem{next: n, system: n.waypoint.GetSystemName()},
		}
	}

	dbTargetWaypoint := database.GetWaypoint(n.waypoint)
	targetData := dbTargetWaypoint.GetData()
	dbWaypoint := database.GetWaypoint(state.Ship.Nav.WaypointSymbol)
	wpData := dbWaypoint.GetData()
	distanceToWaypoint := wpData.GetDistanceFrom(targetData.LimitedWaypointData)
	fuelToWaypoint := util.GetFuelCost(distanceToWaypoint, state.Ship.Nav.FlightMode)

	if fuelToWaypoint > state.Ship.Fuel.Current {
		fuelMarkets := database.GetMarketsSelling([]string{"FUEL"})

		combinedDistances := make(map[entity.Waypoint]int)
		eligibleMarkets := make([]database.MarketRates, 0)
		for _, fuelMarket := range fuelMarkets {
			// We can't go here because it's in a different system
			if fuelMarket.Waypoint.GetSystemName() != state.Ship.Nav.SystemSymbol {
				continue
			}

			fuelMarketWaypoint := fuelMarket.GetLimitedWaypointData()
			distanceToFuelMarket := wpData.GetDistanceFrom(fuelMarketWaypoint)
			fuelToFuelMarket := util.GetFuelCost(distanceToFuelMarket, state.Ship.Nav.FlightMode)
			// We can't go here because it'll take more fuel than we have
			if fuelToFuelMarket > state.Ship.Fuel.Current {
				continue
			}

			distanceFromFuelMarketToTarget := fuelMarketWaypoint.GetDistanceFrom(targetData.LimitedWaypointData)

			// We can't go here because the market is further away then we currently are
			if distanceFromFuelMarketToTarget > distanceToWaypoint {
				continue
			}

			fuelToTarget := util.GetFuelCost(distanceFromFuelMarketToTarget, state.Ship.Nav.FlightMode)
			// We can't go here because going from here to the target would take more fuel than available
			if fuelToTarget > state.Ship.Fuel.Capacity {
				continue
			}
			eligibleMarkets = append(eligibleMarkets, fuelMarket)
			// TODO: should we take into account fuel cost here?
			combinedDistances[fuelMarket.Waypoint] = distanceToFuelMarket + distanceFromFuelMarketToTarget
		}

		if len(eligibleMarkets) == 0 {
			if state.Ship.Nav.FlightMode == "DRIFT" {
				return RoutineResult{
					SetRoutine: Refuel{next: n},
				}
			}
			state.Log("Trying again in drift mode")
			state.Ship.SetFlightMode(state.Context, "DRIFT")
			return RoutineResult{}
		}

		sort.Slice(eligibleMarkets, func(i, j int) bool {
			return combinedDistances[eligibleMarkets[i].Waypoint] < combinedDistances[eligibleMarkets[j].Waypoint]
		})

		state.Log(fmt.Sprintf("Taking a detour via market %s", eligibleMarkets[0].Waypoint))
		return RoutineResult{
			SetRoutine: NavigateTo{waypoint: eligibleMarkets[0].Waypoint, next: Refuel{next: n}},
		}
	}

	_ = state.Ship.EnsureNavState(state.Context, entity.NavOrbit)

	_, err := state.Ship.Navigate(state.Context, n.waypoint)
	if err != nil {
		switch err.Code {
		case http.ErrCooldown:
			return RoutineResult{
				WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
			}
		case http.ErrInsufficientFuelForNav:
			state.Log(err.Message)
			state.Log("Refuelling and trying again")
			if n.nextIfNoFuel != nil {
				return RoutineResult{
					SetRoutine: Refuel{next: n.nextIfNoFuel},
				}
			}
			return RoutineResult{
				SetRoutine: Refuel{next: n},
			}
		case http.ErrShipInTransit, http.ErrNavigateInTransit:
			state.Log("Ship in transit")
			return RoutineResult{
				WaitSeconds: 30,
			}
		case http.ErrShipAtDestination:
			state.Log("Oh we're already there")
			return RoutineResult{
				SetRoutine: n.next,
			}
		}
		state.Log(fmt.Sprintf("Unknown error: %s", err))
		return RoutineResult{
			WaitSeconds: 10,
		}
	}
	state.Log(fmt.Sprintf("Navigating until %s", &state.Ship.Nav.Route.Arrival))
	return RoutineResult{
		WaitUntil:  &state.Ship.Nav.Route.Arrival,
		SetRoutine: n.next,
	}
}

func (n NavigateTo) Name() string {
	return fmt.Sprintf("Navigate to %s -> %s", n.waypoint, n.next.Name())
}
