package routine

import (
	"fmt"
	"math"
	"sort"
	"spacetraders/entity"
	"spacetraders/http"
	"spacetraders/util"
	"strings"
	"time"
)

type Refine struct {
}

func (r Refine) Run(state *State) RoutineResult {

	shipWaypoints := make(map[entity.Waypoint]int)
	for _, state := range *state.States {
		if state.Ship.Registration.Role == "EXCAVATOR" {
			shipWaypoints[state.Ship.Nav.WaypointSymbol]++
		}
	}

	mostShips := entity.Waypoint("")
	for system, amount := range shipWaypoints {
		if shipWaypoints[mostShips] < amount {
			mostShips = system
		}
	}
	state.Log(fmt.Sprintf("Most ships are in %s", mostShips))

	if mostShips != state.Ship.Nav.WaypointSymbol {
		return RoutineResult{SetRoutine: NavigateTo{waypoint: mostShips, next: r}}
	}

	state.WaitingForHttp = true
	_ = state.Ship.EnsureNavState(entity.NavOrbit)
	state.WaitingForHttp = false

	refineableSlots := make([]entity.ShipInventorySlot, 0)
	for _, item := range state.Ship.Cargo.Inventory {
		if util.IsRefineable(item.Symbol) && item.Units > 30 {
			refineableSlots = append(refineableSlots, item)
		}
	}

	if len(refineableSlots) == 0 && state.Ship.Cargo.IsFull() {
		state.Log("Offloading cargo")
		return RoutineResult{SetRoutine: SellExcessInventory{next: r}}
	}

	var cooldown time.Time
	if len(refineableSlots) > 0 {
		sort.Slice(refineableSlots, func(i, j int) bool {
			return refineableSlots[i].Units > refineableSlots[j].Units
		})

		refineTarget := refineableSlots[0]
		state.Log(fmt.Sprintf("Refining %dx %s", refineTarget.Units, refineTarget.Symbol))

		refineResult, err := state.Ship.Refine(refineTarget.Symbol[:strings.Index(refineTarget.Symbol, "_")])

		if err != nil {
			switch err.Code {
			case http.ErrCooldown:
				state.Log("We are on cooldown from a previous running routine")
				return RoutineResult{
					WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
				}
			case http.ErrCargoUnitCountError:
				state.Log(err.Error())
				state.Log("Uhhh cargo count error? why")
				return RoutineResult{}
			}
			return RoutineResult{
				Stop:       true,
				StopReason: err.Error(),
			}
		}
		state.Log(fmt.Sprintf("Refined %v into %v", refineResult.Consumed, refineResult.Produced))

		cooldown = refineResult.Cooldown.Expiration
	}

	state.WaitingForHttp = true
	cargo, _ := state.Ship.GetCargo()
	state.WaitingForHttp = false
	availableSpace := cargo.GetRemainingCapacity()
	usedSpace := 0

	if availableSpace > 0 {
		for _, otherState := range *state.States {
			if otherState.Ship.Registration.Role != "EXCAVATOR" || otherState.Ship.Nav.WaypointSymbol != state.Ship.Nav.WaypointSymbol || otherState.Ship.Nav.Status != state.Ship.Nav.Status {
				continue
			}

			for _, slot := range otherState.Ship.Cargo.Inventory {
				if util.IsRefineable(slot.Symbol) {
					transferAmount := int(math.Min(float64(availableSpace-usedSpace), float64(slot.Units)))
					if transferAmount <= 0 {
						break
					}
					state.Log(fmt.Sprintf("Transferred %dx %s to refinery from %s", transferAmount, slot.Symbol, otherState.Ship.Symbol))
					err := otherState.Ship.TransferCargo(state.Ship.Symbol, slot.Symbol, transferAmount)
					if err != nil {
						state.Log(err.Error())
						break
					}
					usedSpace += transferAmount
				}
			}
		}
	}

	return RoutineResult{
		WaitUntil: &cooldown,
	}

}

func (r Refine) Name() string {
	return "Refine"
}
