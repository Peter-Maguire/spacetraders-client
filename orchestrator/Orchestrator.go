package orchestrator

import (
    "fmt"
    "os"
    "spacetraders/database"
    "spacetraders/entity"
    "spacetraders/routine"
    "spacetraders/ui"
    "time"
)

type Orchestrator struct {
    States       []*routine.State
    Agent        *entity.Agent
    Contract     *entity.Contract
    Haulers      []*entity.Ship
    Channel      chan routine.OrchestratorEvent
    CreditTarget int
    ShipToBuy    string
    Shipyard     entity.Waypoint
}

func Init() *Orchestrator {
    agent, err := entity.GetAgent()

    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    contracts, _ := agent.Contracts()

    var contract *entity.Contract

    for _, c := range *contracts {
        if !c.Fulfilled && c.Accepted {
            contract = &c
            break
        }
    }

    if contract == nil {
        ui.MainLog("No current Contract")
    }

    orc := Orchestrator{
        Agent:        agent,
        Contract:     contract,
        Channel:      make(chan routine.OrchestratorEvent),
        CreditTarget: 300000,
    }

    waypoints, _ := agent.Headquarters.GetSystemWaypoints()

    for _, waypoint := range *waypoints {
        if waypoint.HasTrait("SHIPYARD") {
            ui.MainLog(fmt.Sprintln("Found Shipyard at ", waypoint.Symbol))
            orc.Shipyard = waypoint.Symbol
            break
        }
    }

    ships, _ := orc.Agent.Ships()

    for _, ship := range *ships {
        if ship.Registration.Role == "HAULER" {
            ui.MainLog(fmt.Sprintf("Hauler is %s", ship.Registration))
            orc.Haulers = append(orc.Haulers, &ship)
            break
        }
    }

    shipCount := len(*ships)

    ui.MainLog(fmt.Sprintf("We have %d ships", shipCount))

    if shipCount < 10 {
        orc.ShipToBuy = "SHIP_MINING_DRONE"
    } else if shipCount > 30 {
        orc.ShipToBuy = "SHIP_ORE_HOUND"
    } else {
        orc.ShipToBuy = "SHIP_REFINING_FREIGHTER"
    }

    shipyardStock, err := orc.Shipyard.GetShipyard()
    if err == nil {
        go database.StoreShipCosts(shipyardStock)
        for _, stock := range shipyardStock.Ships {
            if stock.Name == orc.ShipToBuy {
                ui.MainLog(fmt.Sprintf("Ship %s is available to buy at %s for %d credits\n", orc.ShipToBuy, orc.Shipyard, stock.PurchasePrice))
                orc.CreditTarget = stock.PurchasePrice
            }
        }
    }

    go orc.runEvents()

    orc.States = make([]*routine.State, len(*ships))

    ui.MainLog(fmt.Sprint("Starting Routines"))
    for i, ship := range *ships {
        shipPtr := ship
        state := routine.State{
            Agent:    agent,
            Contract: contract,
            Ship:     &shipPtr,
            Haulers:  orc.Haulers,
            EventBus: orc.Channel,
        }
        orc.States[i] = &state

        go orc.routineLoop(&state)
    }
    return &orc
}

func (o *Orchestrator) runEvents() {
    for {
        event := <-o.Channel
        switch event.Name {
        case "sellComplete":
            agent := event.Data.(*entity.Agent)
            //if agent.Credits >= o.CreditTarget && o.Shipyard != "" {
            //    result, err := agent.BuyShip(o.Shipyard, o.ShipToBuy)
            //    if err == nil && result != nil {
            //        ui.MainLog(fmt.Sprintln(result))
            //        state := routine.State{
            //            Contract: o.Contract,
            //            Ship:     result.Ship,
            //            EventBus: o.Channel,
            //        }
            //        ui.MainLog(fmt.Sprintln("New ship", result.Ship.Symbol))
            //        o.States = append(o.States, &state)
            //        go o.routineLoop(&state)
            //    } else {
            //        if err.Data != nil && err.Data["creditsNeeded"] != nil {
            //            o.CreditTarget = int(err.Data["creditsNeeded"].(float64))
            //            ui.MainLog(fmt.Sprintln("Need ", o.CreditTarget))
            //        }
            //        ui.MainLog(fmt.Sprintln("Purchase error", err))
            //    }
            //}
            ui.MainLog(fmt.Sprintf("Credits now: %d/%d\n", agent.Credits, o.CreditTarget))
        case "goodSurveyFound":
            ui.MainLog("Someone found a good survey\n")
            for _, state := range o.States {
                if state.Ship.IsMiningShip() && state.Survey == nil {
                    state.Survey = event.Data.(*entity.Survey)
                }
            }
        case "surveyExhausted":
            ui.MainLog("Survey bad\n")
            for _, state := range o.States {
                if state.Survey == event.Data.(*entity.Survey) {
                    state.Survey = nil
                }
            }
        case "contractComplete":
            ui.MainLog("Contract completed")
            contracts, err := o.Agent.Contracts()
            if err != nil {
                ui.MainLog("Contract get error " + err.Error())
                os.Exit(1)
            }
            for _, c := range *contracts {
                if c.Accepted == false {
                    err = c.Accept()
                    if err != nil {
                        ui.MainLog("Contract accept error " + err.Error())
                        os.Exit(1)
                    } else {
                        ui.MainLog("Accepted new Contract " + c.Id)
                        for _, state := range o.States {
                            state.Contract = &c
                        }
                    }
                }
            }
        }
    }
}

func (o *Orchestrator) routineLoop(state *routine.State) {
    state.CurrentRoutine = routine.DetermineObjective{}
    for {
        routineResult := state.CurrentRoutine.Run(state)
        state.WaitingForHttp = false
        if routineResult.WaitSeconds > 0 {
            state.Log(fmt.Sprintf("Waiting for %d seconds", routineResult.WaitSeconds))
            sleepTime := time.Duration(routineResult.WaitSeconds) * time.Second
            asleepUntil := time.Now().Add(sleepTime)
            state.AsleepUntil = &asleepUntil
            time.Sleep(sleepTime)
        }

        if routineResult.WaitUntil != nil {
            waitTime := routineResult.WaitUntil.Sub(time.Now())
            state.Log(fmt.Sprintf("Waiting until %s (%.f seconds)", routineResult.WaitUntil, waitTime.Seconds()))
            state.AsleepUntil = routineResult.WaitUntil
            time.Sleep(waitTime)
        }
        state.AsleepUntil = nil

        if state.ForceRoutine != nil {
            state.Log("Forced routine change")
            state.CurrentRoutine = state.ForceRoutine
            state.ForceRoutine = nil
            continue
        }

        if routineResult.SetRoutine != nil {
            state.Log(fmt.Sprintf("%s -> %s", state.CurrentRoutine.Name(), routineResult.SetRoutine.Name()))
            state.CurrentRoutine = routineResult.SetRoutine
        }

        if routineResult.Stop {
            state.CurrentRoutine = nil
            state.Log("Stopping Routine")
            break
        }
    }
    state.CurrentRoutine = nil
    state.Log("!!!! Loop exited!")
}
