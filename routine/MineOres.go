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
	next           Routine
}

func (m MineOres) Run(state *State) RoutineResult {

	if !state.Ship.IsMiningShip() {
		state.Log("We shouldnt've got here")
		return RoutineResult{
			SetRoutine: DetermineObjective{},
		}
	}

	if m.latentCooldown != nil && m.latentCooldown.After(time.Now()) {
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
				SetRoutine: GoToMiningArea{},
			}
		case http.ErrShipSurveyExhausted, http.ErrShipSurveyVerification, http.ErrShipSurveyExpired:
			state.Log("Something went wrong with the survey " + err.Error())
			state.FireEvent("surveyExhausted", state.Survey)
			state.Survey = nil
			return RoutineResult{}
		case http.ErrOverExtracted:
			state.Log("Asteroid Over-extracted")
			return RoutineResult{
				SetRoutine: GoToMiningArea{blacklist: []entity.Waypoint{state.Ship.Nav.WaypointSymbol}},
			}
		}

		state.Log(fmt.Sprintf("Unknown error: %s", err))
		// No idea
		return RoutineResult{
			WaitSeconds: 10,
		}
	}

	for _, event := range result.Events {
		state.Log(fmt.Sprintf("!!! Mining Event - %s: %s", event.Name, event.Description))
	}

	mined.WithLabelValues(result.Extraction.Yield.Symbol).Add(float64(result.Extraction.Yield.Units))

	state.Log(fmt.Sprintf("Mined %d %s, cooldown for %d seconds", result.Extraction.Yield.Units, result.Extraction.Yield.Symbol, result.Cooldown.RemainingSeconds))

	if state.ConstructionSite != nil {
		for _, material := range state.ConstructionSite.Materials {
			materialItemSlot := state.Ship.Cargo.GetSlotWithItem(material.TradeSymbol)
			if materialItemSlot != nil && materialItemSlot.Units >= material.GetRemaining() {
				return RoutineResult{
					SetRoutine: DeliverConstructionSiteItem{next: GoToMiningArea{next: m}},
				}
			}
		}
	}

	if state.Contract != nil {
		for _, deliverable := range state.Contract.Terms.Deliver {
			contractItemSlot := state.Ship.Cargo.GetSlotWithItem(deliverable.TradeSymbol)
			if contractItemSlot != nil {
				if contractItemSlot.Units >= deliverable.GetRemaining() {
					state.Log("We have a contract item to deliver")
					return RoutineResult{
						SetRoutine: DeliverContractItem{item: deliverable.TradeSymbol, next: GoToMiningArea{next: m}},
					}
				}

				// TODO: Pre-emptively deliver if all ships combined have enough to finish the procurement
				//total := state.GetTotalOfItemAcrossAllShips(deliverable.TradeSymbol)
				//if total >= deliverable.GetRemaining() && state.GetShipsWithRoleAtOrGoingToWaypoint() {
				//
				//}
			}
		}

	}

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

	if m.next != nil {
		return RoutineResult{
			SetRoutine: m.next,
		}
	}

	return RoutineResult{
		WaitUntil: &result.Cooldown.Expiration,
	}
}

// TODO: replace this with a better system
var uselessItems = []string{}

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
