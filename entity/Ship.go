package entity

import (
    "errors"
    "fmt"
    "spacetraders/http"
    "strings"
    "time"
)

type Ship struct {
    Symbol       string           `json:"symbol"`
    Nav          ShipNav          `json:"nav"`
    Crew         ShipCrew         `json:"crew"`
    Fuel         ShipFuel         `json:"fuel"`
    Frame        ShipFrame        `json:"frame"`
    Reactor      ShipReactor      `json:"reactor"`
    Engine       ShipEngine       `json:"engine"`
    Modules      []ShipMount      `json:"modules"`
    Mounts       []ShipMount      `json:"mounts"`
    Registration ShipRegistration `json:"registration"`
    Cargo        ShipCargo        `json:"cargo"`
}

func (s *Ship) HasMount(mountSymbol string) bool {
    for _, mount := range s.Mounts {
        if mount.Symbol == mountSymbol {
            return true
        }
    }
    return false
}

func (s *Ship) IsMiningShip() bool {
    for _, mount := range s.Mounts {
        if strings.HasPrefix(mount.Symbol, "MOUNT_MINING_LASER") {
            return true
        }
    }
    return false
}

func (s *Ship) Navigate(waypoint Waypoint) (*ShipNav, *http.HttpError) {
    shipUpdate, err := http.Request[Ship]("POST", fmt.Sprintf("my/ships/%s/navigate", s.Symbol), NavigateRequest{
        WaypointSymbol: waypoint,
    })
    if err == nil {
        s.Nav = shipUpdate.Nav
        s.Fuel = shipUpdate.Fuel
        return &shipUpdate.Nav, err
    }
    return nil, err
}

func (s *Ship) Dock() error {
    shipNavUpdate, err := http.Request[Ship]("POST", fmt.Sprintf("my/ships/%s/dock", s.Symbol), nil)
    if shipNavUpdate != nil {
        s.Nav = shipNavUpdate.Nav
    }
    return err
}

func (s *Ship) Refuel() *http.HttpError {
    shipRefuelUpdate, err := http.Request[Ship]("POST", fmt.Sprintf("my/ships/%s/refuel", s.Symbol), nil)
    if shipRefuelUpdate != nil {
        s.Fuel = shipRefuelUpdate.Fuel
    }
    return err
}

func (s *Ship) Orbit() error {
    shipNavUpdate, err := http.Request[Ship]("POST", fmt.Sprintf("my/ships/%s/orbit", s.Symbol), nil)
    if shipNavUpdate != nil {
        s.Nav = shipNavUpdate.Nav
    }
    return err
}

func (s *Ship) SetFlightMode(mode string) *http.HttpError {
    shipNavUpdate, err := http.Request[ShipNav]("PATCH", fmt.Sprintf("my/ships/%s/nav", s.Symbol), map[string]string{
        "flightMode": mode,
    })
    if shipNavUpdate != nil {
        s.Nav = *shipNavUpdate
    }
    return err
}

func (s *Ship) EnsureNavState(state NavState) error {
    if s.Nav.Status == state {
        return nil
    }
    switch state {
    case NavOrbit:
        return s.Orbit()
    case NavDocked:
        return s.Dock()
    }
    return errors.New("invalid nav state")
}

func (s *Ship) SellCargo(cargoSymbol string, units int) (*SellResult, *http.HttpError) {
    sellResult, err := http.Request[SellResult]("POST", fmt.Sprintf("my/ships/%s/sell", s.Symbol), SellRequest{
        Symbol: cargoSymbol,
        Units:  units,
    })
    if sellResult != nil {
        s.Cargo = sellResult.Cargo
    }
    return sellResult, err
}

func (s *Ship) Extract() (*ExtractionResult, *http.HttpError) {
    extractionResult, err := http.Request[ExtractionResult]("POST", fmt.Sprintf("my/ships/%s/extract", s.Symbol), nil)
    if extractionResult != nil {
        s.Cargo = extractionResult.Cargo
    }
    return extractionResult, err
}

func (s *Ship) ExtractSurvey(survey *Survey) (*ExtractionResult, *http.HttpError) {
    extractionResult, err := http.Request[ExtractionResult]("POST", fmt.Sprintf("my/ships/%s/extract", s.Symbol), map[string]*Survey{
        "survey": survey,
    })
    if extractionResult != nil {
        s.Cargo = extractionResult.Cargo
    }
    return extractionResult, err
}

func (s *Ship) Survey() (*SurveyResult, *http.HttpError) {
    return http.Request[SurveyResult]("POST", fmt.Sprintf("my/ships/%s/survey", s.Symbol), nil)
}

func (s *Ship) Jump(system string) (*ShipJumpResult, error) {
    jumpResult, err := http.Request[ShipJumpResult]("POST", fmt.Sprintf("my/ships/%s/jump", s.Symbol), map[string]string{
        "systemSymbol": system,
    })
    if err == nil {
        s.Nav = jumpResult.Nav
        return jumpResult, nil
    }
    return jumpResult, err
}

func (s *Ship) Warp(waypoint Waypoint) (*ShipWarpResult, *http.HttpError) {
    warpResult, err := http.Request[ShipWarpResult]("POST", fmt.Sprintf("my/ships/%s/warp", s.Symbol), map[string]Waypoint{
        "waypointSymbol": waypoint,
    })
    if err == nil {
        s.Fuel = warpResult.Fuel
        s.Nav = warpResult.Nav
    }
    return warpResult, err
}

func (s *Ship) Chart() (*ShipChartResult, *http.HttpError) {
    return http.Request[ShipChartResult]("POST", fmt.Sprintf("my/ships/%s/chart", s.Symbol), nil)
}

func (s *Ship) ScanWaypoints() (*ShipScanWaypointsResult, *http.HttpError) {
    return http.Request[ShipScanWaypointsResult]("POST", fmt.Sprintf("my/ships/%s/scan/waypoints", s.Symbol), nil)
}

func (s *Ship) ScanSystems() (*ShipScanSystemsResult, *http.HttpError) {
    return http.Request[ShipScanSystemsResult]("POST", fmt.Sprintf("my/ships/%s/scan/systems", s.Symbol), nil)
}

func (s *Ship) JettisonCargo(symbol string, amount int) *http.HttpError {
    shipUpdate, err := http.Request[Ship]("POST", fmt.Sprintf("my/ships/%s/jettison", s.Symbol), map[string]any{
        "symbol": symbol,
        "units":  amount,
    })

    if err != nil {
        s.Cargo = shipUpdate.Cargo
    }
    return err
}

func (s *Ship) TransferCargo(ship string, symbol string, amount int) *http.HttpError {
    shipUpdate, err := http.Request[Ship]("POST", fmt.Sprintf("my/ships/%s/transfer", s.Symbol), map[string]any{
        "tradeSymbol": symbol,
        "shipSymbol":  ship,
        "units":       amount,
    })

    if err == nil {
        s.Cargo = shipUpdate.Cargo
    }
    return err
}

func (s *Ship) GetCargo() (*ShipCargo, *http.HttpError) {
    cargo, err := http.Request[ShipCargo]("GET", fmt.Sprintf("my/ships/%s/cargo", s.Symbol), nil)

    s.Cargo = *cargo

    return cargo, err
}

func (s *Ship) Purchase(symbol string, amount int) (*ItemPurchaseResult, *http.HttpError) {
    result, err := http.Request[ItemPurchaseResult]("POST", fmt.Sprintf("my/ships/%s/cargo", s.Symbol), map[string]any{
        "symbol": symbol,
        "units":  amount,
    })

    s.Cargo = *result.Cargo

    return result, err
}

func (s *Ship) NegotiateContract() (*Contract, *http.HttpError) {
    result, err := http.Request[map[string]*Contract]("POST", fmt.Sprintf("my/ships/%s/negotiate/contract", s.Symbol), nil)

    if err != nil {
        return nil, err
    }

    return (*result)["contract"], err
}

type ShipNav struct {
    SystemSymbol   string   `json:"systemSymbol"`
    WaypointSymbol Waypoint `json:"waypointSymbol"`
    Route          NavRoute `json:"route"`
    Status         NavState `json:"status"`
    FlightMode     string   `json:"flightMode"`
}

type NavState string

const (
    NavDocked NavState = "DOCKED"
    NavOrbit  NavState = "IN_ORBIT"
)

type NavRoute struct {
    Departure     WaypointData `json:"departure"`
    Destination   WaypointData `json:"destination"`
    Arrival       time.Time    `json:"arrival"`
    DepartureTime time.Time    `json:"departureTime"`
}

type ShipCrew struct {
    Current  int    `json:"current"`
    Capacity int    `json:"capacity"`
    Required int    `json:"required"`
    Rotation string `json:"rotation"`
    Morale   int    `json:"morale"`
    Wages    int    `json:"wages"`
}

type ShipFuel struct {
    Current  int `json:"current"`
    Capacity int `json:"capacity"`
    Consumed struct {
        Amount    int       `json:"amount"`
        Timestamp time.Time `json:"timestamp"`
    } `json:"consumed"`
}

type ShipFrame struct {
    Symbol         string          `json:"symbol"`
    Name           string          `json:"name"`
    Description    string          `json:"description"`
    ModuleSlots    int             `json:"moduleSlots"`
    MountingPoints int             `json:"mountingPoints"`
    FuelCapacity   int             `json:"fuelCapacity"`
    Condition      int             `json:"condition"`
    Requirements   ShipRequirement `json:"requirements"`
}

type ShipReactor struct {
    Symbol       string          `json:"symbol"`
    Name         string          `json:"name"`
    Description  string          `json:"description"`
    Condition    int             `json:"condition"`
    PowerOutput  int             `json:"powerOutput"`
    Requirements ShipRequirement `json:"requirements"`
}

type ShipEngine struct {
    Symbol       string          `json:"symbol"`
    Name         string          `json:"name"`
    Description  string          `json:"description"`
    Condition    int             `json:"condition"`
    Speed        int             `json:"speed"`
    Requirements ShipRequirement `json:"requirements"`
}

type ShipModule struct {
    Symbol       string          `json:"symbol"`
    Name         string          `json:"name"`
    Description  string          `json:"description"`
    Capacity     int             `json:"capacity,omitempty"`
    Requirements ShipRequirement `json:"requirements"`
    Range        int             `json:"range,omitempty"`
}

type ShipMount struct {
    Symbol       string          `json:"symbol"`
    Name         string          `json:"name"`
    Description  string          `json:"description"`
    Strength     int             `json:"strength"`
    Requirements ShipRequirement `json:"requirements"`
    Deposits     []string        `json:"deposits,omitempty"`
}

type ShipRequirement struct {
    Crew  int `json:"crew"`
    Power int `json:"power"`
    Slots int `json:"slots"`
}

type ShipRegistration struct {
    Name          string `json:"name"`
    FactionSymbol string `json:"factionSymbol"`
    Role          string `json:"role"`
}
type ShipCargo struct {
    Capacity  int                 `json:"capacity"`
    Units     int                 `json:"units"`
    Inventory []ShipInventorySlot `json:"inventory"`
}

func (sc *ShipCargo) GetSlotWithItem(itemSymbol string) *ShipInventorySlot {
    for _, slot := range sc.Inventory {
        if slot.Symbol == itemSymbol {
            return &slot
        }
    }
    return nil
}

type ShipInventorySlot struct {
    Symbol      string `json:"symbol"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Units       int    `json:"units"`
}
