package routine

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"spacetraders/entity"
	"spacetraders/http"
	"time"
)

var (
	mined = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "st_amount_mined",
		Help: "Amount Mined",
	}, []string{"symbol"})
)

type MineOres struct {
}

func (m MineOres) Run(state *State) RoutineResult {

	if state.Ship.Cargo.IsFull() {
		return RoutineResult{
			SetRoutine: FullWait{},
		}
	}

	state.WaitingForHttp = true
	_ = state.Ship.EnsureNavState(entity.NavOrbit)
	state.WaitingForHttp = false

	var result *entity.ExtractionResult
	var err *http.HttpError

	state.WaitingForHttp = true
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

	state.WaitingForHttp = false

	if err != nil {
		switch err.Code {
		case http.ErrCooldown:
			state.Log("We are on cooldown from a previous running routine")
			return RoutineResult{
				WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
			}
		case http.ErrCargoFull:
			hasJettisoned := false
			state.Log("Cargo is full")
			//for _, slot := range state.Ship.Cargo.Inventory {
			//    if m.IsUseless(slot.Symbol) {
			//        state.Log(fmt.Sprintf("Jettison %dx %s", slot.Units, slot.Symbol))
			//        err = state.Ship.JettisonCargo(slot.Symbol, slot.Units)
			//        hasJettisoned = hasJettisoned || err == nil
			//    }
			//}
			//return RoutineResult{WaitSeconds: 10}
			if hasJettisoned {
				return RoutineResult{}
			}
			return RoutineResult{
				WaitSeconds: 30,
				SetRoutine:  FullWait{},
			}
		case http.ErrCannotExtractHere:
			state.Log("We're not at an asteroid field")
			return RoutineResult{
				SetRoutine: GoToMiningArea{GetSurvey{}},
			}
		case http.ErrShipSurveyExhausted, http.ErrShipSurveyVerification, http.ErrShipSurveyExpired:
			state.Log("Something went wrong with the survey " + err.Error())
			state.FireEvent("surveyExhausted", state.Survey)
			state.Survey = nil
			return RoutineResult{}
		}

		state.Log(fmt.Sprintf("Unknown error: %s", err))
		// No idea
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	mined.WithLabelValues(result.Extraction.Yield.Symbol).Add(float64(result.Extraction.Yield.Units))

	state.Log(fmt.Sprintf("Mined %d %s, cooldown for %d seconds", result.Extraction.Yield.Units, result.Extraction.Yield.Symbol, result.Cooldown.RemainingSeconds))

	//if state.Contract == nil {
	//    for _, slot := range result.Cargo.Inventory {
	//        if m.IsUseless(slot.Symbol) {
	//            state.Log(fmt.Sprintf("Jettison %dx %s", slot.Units, slot.Symbol))
	//            err = state.Ship.JettisonCargo(slot.Symbol, slot.Units)
	//        }
	//    }
	//}

	return RoutineResult{
		WaitUntil: &result.Cooldown.Expiration,
	}
}

var uselessItems = []string{"QUARTZ_SAND", "ICE_WATER", ""}

func (m *MineOres) IsUseless(item string) bool {
	for _, uselessItem := range uselessItems {
		if uselessItem == item {
			return true
		}
	}
	return false
}

func (m MineOres) Name() string {
	return "Mine Ores"
}
