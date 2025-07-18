package routine

import (
	"fmt"
	"math"
	"sort"
	"spacetraders/entity"
)

type Haul struct {
}

func (h Haul) Run(state *State) RoutineResult {
	//state.StatesMutex.Lock()

	_ = state.Ship.EnsureNavState(state.Context, entity.NavDocked)

	cargo, _ := state.Ship.GetCargo(state.Context)

	cargoCount := cargo.Units

	sellables := cargo.Inventory

	ships := make([]*entity.Ship, 0)
	haulerCount := len(state.Haulers)
	haulerNum := 0

	for i, hauler := range state.Haulers {
		if hauler.Symbol == state.Ship.Symbol {
			haulerNum = i
			break
		}
	}

	for i, otherState := range *state.States {
		if haulerCount > 1 && (otherState.Ship.Registration.Role != "EXCAVATOR" || i%haulerCount == haulerNum) {
			continue
		}
		//if otherState.Ship.Cargo.Capacity-otherState.Ship.Cargo.Units <= 0 {
		ships = append(ships, otherState.Ship)
		//}
	}

	//sort.Slice(ships, func(i, j int) bool {
	//return ships[i].Cargo.Capacity-ships[i].Cargo.Units > ships[j].Cargo.Capacity-ships[j].Cargo.Units
	//})

	//state.Log(fmt.Sprintf("Most full: %d/%d / Least full: %d/%d", ships[0].Cargo.Capacity-ships[0].Cargo.Units, ships[0].Cargo.Capacity, ships[len(ships)-1].Cargo.Capacity-ships[len(ships)-1].Cargo.Units, ships[len(ships)-1].Cargo.Capacity))

	state.Log(fmt.Sprintf("Current cargo count: %d. #%d of %d haulers. We're in charge of %d ships", cargoCount, haulerNum, haulerCount, len(ships)))

	full := false
	for _, ship := range ships {
		if ship.Registration.Role != "EXCAVATOR" {
			continue
		}

		sort.Slice(ship.Cargo.Inventory, func(i, j int) bool {
			return ship.Cargo.Inventory[i].Units > ship.Cargo.Inventory[j].Units
		})

		state.Log(fmt.Sprintf("Biggest %d / Smallest %d", ship.Cargo.Inventory[0].Units, ship.Cargo.Inventory[len(ship.Cargo.Inventory)-1].Units))

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
			transferAmount := int(math.Min(float64(slot.Units), float64(remainingCapacity)))
			state.Log(fmt.Sprintf("Transferring %d/%d %s from %s to %s (%d/%d cargo)", transferAmount, slot.Units, slot.Symbol, ship.Symbol, state.Ship.Symbol, cargoCount, state.Ship.Cargo.Capacity))
			err := ship.TransferCargo(state.Context, state.Ship.Symbol, slot.Symbol, transferAmount)
			if err != nil {
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
		state.Log("Time to start selling")
		for _, slot := range sellables {
			sellResult, err := state.Ship.SellCargo(state.Context, slot.Symbol, slot.Units)
			if err == nil {
				state.Agent = &sellResult.Agent
				state.Log(fmt.Sprintf("Sold %dx %s for %d credits", slot.Units, slot.Symbol, sellResult.Transaction.TotalPrice))
			} else {
				state.Log(err.Error())
			}
		}
	}

	state.FireEvent("sellComplete", state.Agent)

	if len(sellables) == 0 {
		return RoutineResult{
			WaitSeconds: 5,
		}
	}

	return RoutineResult{}
}

func (h Haul) Name() string {
	return "Haul"
}
