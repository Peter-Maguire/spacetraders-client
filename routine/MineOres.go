package routine

import (
	"fmt"
	"spacetraders/entity"
	"spacetraders/http"
	"time"
)

type MineOres struct {
}

func (m MineOres) Run(state *State) RoutineResult {
	_ = state.Ship.EnsureNavState(entity.NavOrbit)

	var result *entity.ExtractionResult
	var err *http.HttpError

	if state.Survey != nil {
		if state.Survey.Expiration.Before(time.Now()) {
			state.Log("Survey has expired")
			return RoutineResult{
				SetRoutine: GetSurvey{},
			}
		}
		result, err = state.Ship.ExtractSurvey(state.Survey)
	} else {
		result, err = state.Ship.Extract()
	}

	if err != nil {
		switch err.Code {
		case http.ErrCooldown:
			state.Log("We are on cooldown from a previous running routine")
			return RoutineResult{
				WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
			}
		case http.ErrCargoFull:
			state.Log("Cargo is full")
			return RoutineResult{
				SetRoutine: SellExcessInventory{},
			}
		case http.ErrCannotExtractHere:
			state.Log("We're not at an asteroid field")
			return RoutineResult{
				SetRoutine: GoToAsteroidField{},
			}
		case http.ErrShipSurveyExhausted, http.ErrShipSurveyVerification, http.ErrShipSurveyExpired:
			state.Log("Something went wrong with the survey")
			state.Survey = nil
			return RoutineResult{}
		}

		state.Log(fmt.Sprintf("Unknown error: %s", err))
		// No idea
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	state.Log(fmt.Sprintf("Mined %d %s, cooldown for %d seconds", result.Extraction.Yield.Units, result.Extraction.Yield.Symbol, result.Cooldown.RemainingSeconds))

	if result.Cargo.Units >= result.Cargo.Capacity-5 {
		state.Log("Inventory is near to or completely full, time to sell")
		return RoutineResult{
			SetRoutine: SellExcessInventory{},
			WaitUntil:  &result.Cooldown.Expiration,
		}
	}

	return RoutineResult{
		WaitUntil: &result.Cooldown.Expiration,
	}
}

func (m MineOres) Name() string {
	return "Mine Ores"
}
