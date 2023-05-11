package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
	"time"
)

func MineOres(state *entity.State) RoutineResult {
	_ = state.Ship.EnsureNavState(entity.NavOrbit)

	var result *entity.ExtractionResult
	var err *http.HttpError

	if state.Survey != nil {
		if state.Survey.Expiration.Before(time.Now()) {
			fmt.Println("Survey has expired")
			return RoutineResult{
				SetRoutine: GetSurvey,
			}
		}
		result, err = state.Ship.ExtractSurvey(state.Survey)
	} else {
		result, err = state.Ship.Extract()
	}

	if err != nil {
		switch err.Code {
		case http.ErrCooldown:
			fmt.Println("We are on cooldown from a previous running routine")
			return RoutineResult{
				WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
			}
		case http.ErrCargoFull:
			fmt.Println("Cargo is full")
			return RoutineResult{
				SetRoutine: SellExcessInventory,
			}
		}
		fmt.Println("Unknown error", err)
		// No idea
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	fmt.Printf("Mined %d %s, cooldown for %d seconds\n", result.Extraction.Yield.Units, result.Extraction.Yield.Symbol, result.Cooldown.RemainingSeconds)

	if result.Cargo.Units >= result.Cargo.Capacity-5 {
		fmt.Println("Inventory is near to or completely full, time to sell")
		return RoutineResult{
			SetRoutine:  SellExcessInventory,
			WaitSeconds: result.Cooldown.RemainingSeconds,
		}
	}

	return RoutineResult{
		WaitSeconds: result.Cooldown.RemainingSeconds,
	}

}
