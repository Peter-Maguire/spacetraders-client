package routine

import (
	"fmt"
)

type DeliverConstructionSiteItem struct {
	next Routine
}

func (r DeliverConstructionSiteItem) Run(state *State) RoutineResult {

	if state.Ship.Nav.WaypointSymbol != state.ConstructionSite.Symbol {
		state.Log("Going to construction site location")
		return RoutineResult{
			SetRoutine: NavigateTo{
				waypoint: state.ConstructionSite.Symbol,
				next:     r,
			},
		}
	}

	state.Ship.GetCargo(state.Context)

	for _, cargo := range state.Ship.Cargo.Inventory {
		constructionMaterial := state.ConstructionSite.GetMaterial(cargo.Symbol)
		if constructionMaterial == nil {
			continue
		}

		amountToSupply := min(cargo.Units, constructionMaterial.GetRemaining())
		_, err := state.ConstructionSite.Supply(state.Context, state.Ship, cargo.Symbol, amountToSupply)
		if err != nil {
			state.Log(fmt.Sprintf("Supply error: %v", err))
		}

	}

	return RoutineResult{
		SetRoutine: SellExcessInventory{next: r.next},
	}
}

func (r DeliverConstructionSiteItem) Name() string {
	return fmt.Sprintf("Deliver Construction Items -> %s", r.next.Name())
}
