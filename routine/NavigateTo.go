package routine

import (
	"fmt"
	"sort"
	"spacetraders/constant"
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
	isDetour     bool
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
	distanceToTarget := wpData.GetDistanceFrom(targetData.LimitedWaypointData)
	fuelToWaypoint := util.GetFuelCost(distanceToTarget, state.Ship.Nav.FlightMode)
	cruiseFuelToWaypoint := util.GetFuelCost(distanceToTarget, constant.FlightModeCruise)

	if state.Ship.Fuel.Capacity > 0 {
		if cruiseFuelToWaypoint < state.Ship.Fuel.Current {
			state.Log("Setting to cruise as we can get there on our fuel level")
			state.Ship.EnsureFlightMode(state.Context, constant.FlightModeCruise)
		} else if fuelToWaypoint >= state.Ship.Fuel.Capacity {
			state.Log("Setting to drift as we can't get there with our max fuel level")
			state.Ship.EnsureFlightMode(state.Context, constant.FlightModeDrift)
		}
		if !n.isDetour && (fuelToWaypoint > state.Ship.Fuel.Current || state.Ship.Nav.FlightMode == constant.FlightModeDrift) {
			fuelMarkets := database.GetMarketsSelling([]string{"FUEL"})

			combinedDistances := make(map[entity.Waypoint]int)
			eligibleMarkets := make([]database.MarketRates, 0)
			for _, fuelMarket := range fuelMarkets {
				// We can't go here because it's in a different system
				if fuelMarket.Waypoint.GetSystemName() != state.Ship.Nav.SystemSymbol {
					//fmt.Printf("NAV: Discounting %s because it's not in our system (%s)\n", fuelMarket.Waypoint, state.Ship.Nav.SystemSymbol)
					continue
				}

				if fuelMarket.Waypoint == state.Ship.Nav.WaypointSymbol {
					//fmt.Printf("NAV: Discounting %s as it's the waypoint we are currently at\n", fuelMarket.Waypoint)
					continue
				}

				if fuelMarket.Waypoint == n.waypoint {
					//fmt.Printf("NAV: Discounting %s as it's the waypoint we are currently going to\n", fuelMarket.Waypoint)
					continue
				}

				fuelMarketWaypoint := fuelMarket.GetLimitedWaypointData()
				distanceToFuelMarket := wpData.GetDistanceFrom(fuelMarketWaypoint)
				fuelToFuelMarket := util.GetFuelCost(distanceToFuelMarket, state.Ship.Nav.FlightMode)

				if distanceToFuelMarket > distanceToTarget {
					//fmt.Printf("NAV: Discounting %s because it's further away (%d) then the target waypoint (%d)\n", fuelMarket.Waypoint, distanceToFuelMarket, distanceToTarget)
					continue
				}

				// We can't go here because it'll take more fuel than we have
				if fuelToFuelMarket > state.Ship.Fuel.Current {
					//fmt.Printf("NAV: Discounting %s because the fuel cost to get to there (%d) is higher than our current fuel (%d)\n", fuelMarket.Waypoint, fuelToFuelMarket, state.Ship.Fuel.Current)
					continue
				}

				if state.Ship.Nav.FlightMode == constant.FlightModeCruise && fuelToFuelMarket <= 2 {
					//fmt.Printf("NAV: Discounting %s because it is %d fuel away from our current waypoint\n", fuelMarket.Waypoint, fuelToFuelMarket)
				}

				distanceFromFuelMarketToTarget := fuelMarketWaypoint.GetDistanceFrom(targetData.LimitedWaypointData)

				// We can't go here because the market is further away then we currently are
				if distanceFromFuelMarketToTarget >= distanceToTarget {
					//fmt.Printf("NAV: Discounting %s because the distance from the fuel market to the target (%d) is higher than our current distance from the target (%d)\n", fuelMarket.Waypoint, distanceFromFuelMarketToTarget, distanceToTarget)
					continue
				}

				fuelToTarget := util.GetFuelCost(distanceFromFuelMarketToTarget, state.Ship.Nav.FlightMode)
				// We can't go here because going from here to the target would take more fuel than available
				if fuelToTarget > state.Ship.Fuel.Capacity {
					//fmt.Printf("NAV: Discounting %s because the fuel from the market to the target (%d) is higher than our capacity (%d)\n", fuelMarket.Waypoint, fuelToTarget, state.Ship.Fuel.Capacity)
					continue
				}
				fmt.Printf("NAV: Market %s is eligible with distance %d+%d", fuelMarket.Waypoint, distanceToFuelMarket, distanceFromFuelMarketToTarget)
				eligibleMarkets = append(eligibleMarkets, fuelMarket)
				// TODO: should we take into account fuel cost here?
				combinedDistances[fuelMarket.Waypoint] = distanceToFuelMarket + distanceFromFuelMarketToTarget
			}

			if len(eligibleMarkets) == 0 {
				state.Log("No eligible markets found")
				if state.Ship.Nav.FlightMode != "DRIFT" {
					state.Log("Trying again in drift mode")
					state.Ship.SetFlightMode(state.Context, constant.FlightModeDrift)
					return RoutineResult{}
				}
			} else {
				// TODO: shouldn't we go to the *furthest* away we can get on the first leg?
				sort.Slice(eligibleMarkets, func(i, j int) bool {
					return combinedDistances[eligibleMarkets[i].Waypoint] < combinedDistances[eligibleMarkets[j].Waypoint]
				})

				state.Log(fmt.Sprintf("Taking a detour via market %s (distance %d)", eligibleMarkets[0].Waypoint, combinedDistances[eligibleMarkets[0].Waypoint]))

				return RoutineResult{
					SetRoutine: NavigateTo{isDetour: true, waypoint: eligibleMarkets[0].Waypoint, next: Refuel{next: n}},
				}
			}
		}
		// TODO: this shouldn't be duplicated, try and figure out the logic here
		if fuelToWaypoint >= state.Ship.Fuel.Current {
			state.Log("Setting to drift as we can't get there with our current fuel level")
			state.Ship.EnsureFlightMode(state.Context, constant.FlightModeDrift)
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

	state.Ship.EnsureFlightMode(state.Context, "CRUISE")
	state.Log(fmt.Sprintf("Navigating until %s", &state.Ship.Nav.Route.Arrival))
	return RoutineResult{
		WaitUntil:  &state.Ship.Nav.Route.Arrival,
		SetRoutine: n.next,
	}
}

func (n NavigateTo) Name() string {
	return fmt.Sprintf("Navigate to %s -> %s", n.waypoint, n.next.Name())
}
