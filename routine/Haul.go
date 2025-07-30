package routine

import (
	"fmt"
	"sort"
	"spacetraders/constant"
	"spacetraders/entity"
	"spacetraders/http"
	"time"
)

type Haul struct {
}

func (h Haul) Run(state *State) RoutineResult {
	//state.StatesMutex.Lock()

	_ = state.Ship.EnsureNavState(state.Context, entity.NavOrbit)

	cargo, _ := state.Ship.GetCargo(state.Context)

	cargoCount := cargo.Units

	sellables := cargo.Inventory

	ships := make([]*entity.Ship, 0)

	shipsPerWaypoint := make(map[entity.Waypoint]int)
	for _, otherState := range *state.States {
		if otherState.Ship.Registration.Role != constant.ShipRoleExcavator && otherState.Ship.Registration.Role != constant.ShipRoleCommand {
			continue
		}

		if otherState.Ship.Nav.Status == "IN_TRANSIT" {
			shipsPerWaypoint[otherState.Ship.Nav.Route.Destination.Symbol]++
		} else {
			shipsPerWaypoint[otherState.Ship.Nav.WaypointSymbol]++
		}
	}

	var mostShips *entity.Waypoint
	var mostShipsCount int

	for wp, count := range shipsPerWaypoint {
		if count > mostShipsCount {
			mostShipsCount = count
			mostShips = &wp
		}
	}

	if mostShips == nil {
		state.Log("Couldn't find any not in-transit excavators")
		if state.Ship.Cargo.Units > 0 {
			return RoutineResult{
				SetRoutine:  SellExcessInventory{next: h},
				WaitSeconds: 10,
			}
		}
		return RoutineResult{
			WaitSeconds: 60,
		}
	}

	if state.Ship.Nav.WaypointSymbol != *mostShips {
		state.Log(fmt.Sprintf("%dx excavators are at %s", mostShipsCount, *mostShips))
		return RoutineResult{SetRoutine: NavigateTo{waypoint: *mostShips, next: h}}
	}

	for _, otherState := range *state.States {
		// TODO: maybe this should be different, or is it completely redundant now?
		//if otherState.Ship.Cargo.Capacity-otherState.Ship.Cargo.Units <= 0 {
		ships = append(ships, otherState.Ship)
		//}
	}

	//sort.Slice(ships, func(i, j int) bool {
	//return ships[i].Cargo.Capacity-ships[i].Cargo.Units > ships[j].Cargo.Capacity-ships[j].Cargo.Units
	//})

	//state.Log(fmt.Sprintf("Most full: %d/%d / Least full: %d/%d", ships[0].Cargo.Capacity-ships[0].Cargo.Units, ships[0].Cargo.Capacity, ships[len(ships)-1].Cargo.Capacity-ships[len(ships)-1].Cargo.Units, ships[len(ships)-1].Cargo.Capacity))

	//state.Log(fmt.Sprintf("Current cargo count: %d. #%d of %d haulers. We're in charge of %d ships", cargoCount, haulerNum, haulerCount, len(ships)))

	full := false
	for _, ship := range ships {
		if len(ship.Cargo.Inventory) == 0 || !h.ShouldHaulFrom(state, ship) {
			continue
		}

		if state.Ship.Nav.Status == "IN_TRANSIT" {
			continue
		}

		sort.Slice(ship.Cargo.Inventory, func(i, j int) bool {
			return ship.Cargo.Inventory[i].Units > ship.Cargo.Inventory[j].Units
		})

		//state.Log(fmt.Sprintf("Biggest %d / Smallest %d", ship.Cargo.Inventory[0].Units, ship.Cargo.Inventory[len(ship.Cargo.Inventory)-1].Units))

		for _, slot := range ship.Cargo.Inventory {
			if slot.Symbol == "ANTIMATTER" {
				continue
			}
			remainingCapacity := state.Ship.Cargo.Capacity - cargoCount
			if remainingCapacity <= 0 {
				state.Log("Cargo is now full")
				full = true
				break
			}
			transferAmount := min(slot.Units, remainingCapacity)
			state.Log(fmt.Sprintf("Transferring %d/%d %s from %s to %s (%d/%d cargo)", transferAmount, slot.Units, slot.Symbol, ship.Symbol, state.Ship.Symbol, cargoCount, state.Ship.Cargo.Capacity))
			err := ship.TransferCargo(state.Context, state.Ship.Symbol, slot.Symbol, transferAmount)
			if err != nil {
				if err.Code == http.ErrShipInTransit {
					t, _ := time.Parse(time.RFC3339, err.Data["arrival"].(string))
					return RoutineResult{WaitUntil: &t}
				}
				state.Log(err.Error())
				full = true
			} else {
				exists := false
				for i, sellable := range sellables {
					if sellable.Symbol == slot.Symbol {
						exists = true
						sellables[i].Units += transferAmount
						break
					}
				}
				if !exists {
					sellables = append(sellables, entity.ShipInventorySlot{
						Symbol: slot.Symbol,
						Units:  transferAmount,
					})
				}
				cargoCount += transferAmount
			}
			break
		}
		if full {
			break
		}
	}
	//state.StatesMutex.Unlock()

	if full {
		return RoutineResult{
			SetRoutine: Jettison{nextIfFailed: SellExcessInventory{next: h}, nextIfSuccessful: h},
		}
	}

	if len(sellables) == 0 {
		return RoutineResult{
			WaitSeconds: 5,
		}
	}

	allAreWaiting := true
	var lowestWaitTime *time.Time
	for _, otherState := range *state.States {
		if !h.ShouldHaulFrom(state, otherState.Ship) {
			continue
		}

		if otherState.AsleepUntil == nil {
			allAreWaiting = false
			break
		}
		if lowestWaitTime == nil || otherState.AsleepUntil.Before(*lowestWaitTime) {
			lowestWaitTime = otherState.AsleepUntil
		}
	}

	if lowestWaitTime != nil && allAreWaiting {
		tPlusOne := lowestWaitTime.Add(1 * time.Second)
		lowestWaitTime = &tPlusOne
		state.Log("All ships are waiting, so wait until the first one is finished.")
		return RoutineResult{
			WaitUntil: lowestWaitTime,
		}
	}

	return RoutineResult{
		WaitSeconds: 10,
	}
}

func (h Haul) ShouldHaulFrom(state *State, ship *entity.Ship) bool {
	return (ship.Registration.Role == "EXCAVATOR" || ship.Registration.Role == "COMMAND") && ship.Nav.Status != "IN_TRANSIT" && ship.Nav.WaypointSymbol == state.Ship.Nav.WaypointSymbol
}

func (h Haul) Name() string {
	return "Haul"
}
