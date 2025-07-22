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
	latentCooldown *time.Time
}

func (m MineOres) Run(state *State) RoutineResult {

	if m.latentCooldown != nil {
		state.Log("Waiting for Cooldown")
		return RoutineResult{WaitUntil: m.latentCooldown}
	}

	_ = state.Ship.EnsureNavState(state.Context, entity.NavOrbit)

	var result *entity.ExtractionResult
	var err *http.HttpError

	if state.Survey != nil {
		if state.Survey.Expiration.Before(time.Now()) {
			state.Log("Survey has expired")
			return RoutineResult{
				SetRoutine: GetSurvey{},
			}
		}
		result, err = state.Ship.ExtractSurvey(state.Context, state.Survey)
	} else {
		result, err = state.Ship.Extract(state.Context)
	}

	if err != nil {
		switch err.Code {
		case http.ErrCooldown:
			state.Log("We are on cooldown from a previous running routine")
			return RoutineResult{
				WaitSeconds: int(err.Data["cooldown"].(map[string]any)["remainingSeconds"].(float64)),
			}
		case http.ErrCargoFull:
			return RoutineResult{
				SetRoutine: Jettison{
					nextIfFailed:     FullWait{},
					nextIfSuccessful: m,
				},
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

	if state.Ship.Cargo.IsFull() {
		return RoutineResult{
			SetRoutine: Jettison{
				nextIfFailed: FullWait{},
				nextIfSuccessful: MineOres{
					latentCooldown: &result.Cooldown.Expiration,
				},
			},
		}
	}

	return RoutineResult{
		WaitUntil: &result.Cooldown.Expiration,
	}
}

var uselessItems = []string{"QUARTZ_SAND", "ICE_WATER"}

func (m MineOres) IsUseless(item string) bool {
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
