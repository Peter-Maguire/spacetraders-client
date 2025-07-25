package routine

import (
	"fmt"
	"sort"
	"spacetraders/database"
	"spacetraders/entity"
	"strings"
)

type GoToMiningArea struct {
	next      Routine
	blacklist []entity.Waypoint
}

func (g GoToMiningArea) Run(state *State) RoutineResult {
	if g.next == nil {
		g.next = MineOres{}
	}
	_ = state.Ship.EnsureNavState(state.Context, entity.NavOrbit)

	waypointsPtr, _ := state.Ship.Nav.WaypointSymbol.GetSystemWaypoints(state.Context)
	waypoints := *waypointsPtr
	waypointScores := make(map[entity.Waypoint]int)

	//currentWaypoint, _ := state.Ship.Nav.WaypointSymbol.GetWaypointData(state.Context)

	waypointData := make([]*database.Waypoint, len(waypoints))
	for i, waypoint := range waypoints {
		waypointData[i] = database.GetWaypoint(waypoint.Symbol)
	}

	eligibleWaypoints := make([]entity.WaypointData, 0)
	for _, waypoint := range waypoints {
		if g.IsWaypointBlacklisted(waypoint.Symbol) {
			fmt.Printf("%s was marked blacklisted\n", waypoint.Symbol)
			continue
		}
		eligible, score := g.ScoreWaypoint(waypoint, waypointData)
		if !eligible {
			fmt.Printf("%s was marked ineligible\n", waypoint.Symbol)
			continue
		}
		eligibleWaypoints = append(eligibleWaypoints, waypoint)
		waypointScores[waypoint.Symbol] = score
	}

	sort.Slice(eligibleWaypoints, func(i, j int) bool {
		return waypointScores[eligibleWaypoints[i].Symbol] > waypointScores[eligibleWaypoints[j].Symbol]
	})

	if len(waypointScores) == 0 {
		state.Log("No good waypoints found within reach")
		if state.Ship.Fuel.IsFull() {
			if state.Ship.Nav.FlightMode == "DRIFT" {
				if state.Ship.Nav.SystemSymbol != state.Agent.Headquarters.GetSystemName() {
					state.Log("Going back to home system")
					return RoutineResult{
						SetRoutine: GoToSystem{
							system: state.Agent.Headquarters.GetSystemName(),
							next:   g,
						},
					}
				}
				if state.Ship.Cargo.Units > 0 {
					state.Log("Unable to find anywhere to mine")
					return RoutineResult{SetRoutine: SellExcessInventory{next: g}}
				}

				return RoutineResult{
					Stop:       true,
					StopReason: "Unable to find anywhere to mine in range",
				}
			}
			state.Log("Trying again in drift mode")
			state.Ship.SetFlightMode(state.Context, "DRIFT")
			return RoutineResult{}
		}

		state.Log("Attempting to refuel and trying again")
		return RoutineResult{
			SetRoutine: Refuel{
				next: g,
			},
		}
	}

	bestWaypoint := eligibleWaypoints[0]

	state.Log(fmt.Sprintf("Choosing waypoint %s which has score of %d", bestWaypoint.Symbol, waypointScores[bestWaypoint.Symbol]))

	return RoutineResult{SetRoutine: NavigateTo{
		waypoint: bestWaypoint.Symbol,
		next:     g.next,
	}}

	//if state.Ship.Nav.SystemSymbol != state.Agent.Headquarters.GetSystemName() {
	//	return RoutineResult{
	//		SetRoutine: GoToJumpGate{next: GoToSystem{
	//			system: state.Agent.Headquarters.GetSystemName(),
	//			next:   g,
	//		}},
	//	}
	//}
	//
	//state.Log("Couldn't find a waypoint pointing to an asteroid field")
	//return RoutineResult{
	//	WaitSeconds: 60,
	//}
}

func (g GoToMiningArea) ScoreWaypoint(waypoint entity.WaypointData, waypoints []*database.Waypoint) (bool, int) {
	score := 0
	if waypoint.HasTrait("PRECIOUS_METAL_DEPOSITS") {
		score += 15
	}

	if waypoint.HasTrait("RARE_METAL_DEPOSITS") {
		score += 10
	}

	if waypoint.HasTrait("COMMON_METAL_DEPOSITS") {
		score += 5
	}

	if waypoint.HasTrait("MINERAL_DEPOSITS") {
		score += 1
	}

	if score <= 0 {
		return false, score
	}

	if waypoint.HasTrait("MARKETPLACE") {
		score += 1
	}

	if waypoint.HasTrait("OVERCROWDED") {
		return false, 0
	}

	if waypoint.HasTrait("BARREN") {
		return false, 0
	}

	if waypoint.HasTrait("STRIPPED") {
		return false, 0
	}

	if waypoint.HasTrait("UNSTABLE") {
		return false, 0
	}

	if waypoint.HasTrait("CRITICAL_LIMIT") {
		return false, 0
	}

	closestDistance := 5000000
	var closestWaypoint *database.Waypoint
	for _, dbWaypoint := range waypoints {
		if dbWaypoint.Waypoint == string(waypoint.Symbol) {
			continue
		}

		marketData := dbWaypoint.GetMarketData()
		if marketData == nil {
			continue
		}
		buysOres := false
		for _, tg := range marketData.TradeGoods {
			if strings.HasSuffix(tg.Symbol, "_ORE") {
				buysOres = true
				break
			}
		}

		if !buysOres {
			continue
		}

		wpData := dbWaypoint.GetData()
		distance := waypoint.GetDistanceFrom(wpData.LimitedWaypointData)
		if closestWaypoint == nil || distance < closestDistance {
			closestDistance = distance
			closestWaypoint = dbWaypoint
		}
	}

	if closestWaypoint == nil {
		fmt.Printf("No closest waypoint found for %s\n", waypoint.Symbol)
		return true, score - 2000
	}

	score -= closestDistance

	return true, score
}

func (g GoToMiningArea) IsWaypointBlacklisted(waypoint entity.Waypoint) bool {
	if g.blacklist == nil {
		return false
	}
	for _, bl := range g.blacklist {
		if bl == waypoint {
			return true
		}
	}
	return false
}

func (g GoToMiningArea) Name() string {
	if g.next != nil {
		return fmt.Sprintf("Go To Mining Area -> %s", g.next.Name())
	}
	return "Go To Mining Area"
}
