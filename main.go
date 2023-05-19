package main

import (
    "fmt"
    "os"
    "spacetraders/database"
    "spacetraders/http"
    "spacetraders/orchestrator"
    "spacetraders/ui"
    "time"
)

var enableUi = false

var orc *orchestrator.Orchestrator

func main() {
    enableUi = os.Getenv("DISABLE_UI") != "1"
    if enableUi {
        go ui.Init()
    }
    http.Init()
    database.Init()

    orc = orchestrator.Init()

    if enableUi {
        updateShipStates()
    }
}

func updateShipStates() {
    ticker := time.NewTicker(1 * time.Second)
    for {
        <-ticker.C
        output := ""
        //orc.StatesMutex.Lock()
        for _, state := range orc.States {
            if state != nil && state.Ship != nil {
                output += state.Ship.Symbol
                if state.CurrentRoutine != nil {
                    output += fmt.Sprintf(" (%s)", state.CurrentRoutine.Name())
                } else {
                    output += " Stopped"
                }
                if state.WaitingForHttp {
                    output += " Waiting for HTTP"
                }
                if state.AsleepUntil != nil {
                    output += fmt.Sprintf(" Sleeping for %.f seconds", state.AsleepUntil.Sub(time.Now()).Seconds())
                }
                output += "\n"
            }
        }
        //orc.StatesMutex.Unlock()

        httpOutput := fmt.Sprintf("Request Backlog: %d (%d)", len(http.RequestBuffer), http.Waiting)
        if http.IsRunningRequests {
            httpOutput += " (Active)"
        }
        httpOutput += "\n"
        http.RBufferLock.Lock()
        for i, request := range http.RequestBuffer {
            httpOutput += fmt.Sprintf("%d x%d [%d] %s %s %s\n", i+1, len(request.ReturnChannels), request.Priority, request.Req.Method, request.Req.URL.Path, request.Req.URL.Query().Encode())
        }
        http.RBufferLock.Unlock()

        ui.WriteShipState(output, httpOutput)
    }
}
