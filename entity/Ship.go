package entity

import (
	"context"
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

func (s *Ship) CanWarp() bool {
	for _, module := range s.Modules {
		if strings.Contains(module.Symbol, "WARP") {
			return true
		}
	}
	return false
}

func (s *Ship) Navigate(ctx context.Context, waypoint Waypoint) (*ShipNav, *http.HttpError) {
	shipUpdate, err := http.Request[Ship](ctx, "POST", fmt.Sprintf("my/ships/%s/navigate", s.Symbol), NavigateRequest{
		WaypointSymbol: waypoint,
	})
	if err == nil {
		s.Nav = shipUpdate.Nav
		s.Fuel = shipUpdate.Fuel
		return &shipUpdate.Nav, err
	}
	return nil, err
}

func (s *Ship) Dock(ctx context.Context) error {
	shipNavUpdate, err := http.Request[Ship](ctx, "POST", fmt.Sprintf("my/ships/%s/dock", s.Symbol), nil)
	if shipNavUpdate != nil {
		s.Nav = shipNavUpdate.Nav
	}
	return err
}

func (s *Ship) Refuel(ctx context.Context) *http.HttpError {
	shipRefuelUpdate, err := http.Request[Ship](ctx, "POST", fmt.Sprintf("my/ships/%s/refuel", s.Symbol), nil)
	if shipRefuelUpdate != nil {
		s.Fuel = shipRefuelUpdate.Fuel
	}
	return err
}

func (s *Ship) Orbit(ctx context.Context) error {
	shipNavUpdate, err := http.Request[Ship](ctx, "POST", fmt.Sprintf("my/ships/%s/orbit", s.Symbol), nil)
	if shipNavUpdate != nil {
		s.Nav = shipNavUpdate.Nav
	}
	return err
}

func (s *Ship) SetFlightMode(ctx context.Context, mode string) *http.HttpError {
	shipNavUpdate, err := http.Request[Ship](ctx, "PATCH", fmt.Sprintf("my/ships/%s/nav", s.Symbol), map[string]string{
		"flightMode": mode,
	})
	if shipNavUpdate != nil {
		s.Nav = shipNavUpdate.Nav
	}
	return err
}

func (s *Ship) EnsureNavState(ctx context.Context, state NavState) error {
	if s.Nav.Status == state {
		return nil
	}
	var err error
	switch state {
	case NavOrbit:
		err = s.Orbit(ctx)
	case NavDocked:
		err = s.Dock(ctx)
	default:
		err = errors.New("unknown nav state")
	}
	return err
}

func (s *Ship) SellCargo(ctx context.Context, cargoSymbol string, units int) (*SellResult, *http.HttpError) {
	sellResult, err := http.Request[SellResult](ctx, "POST", fmt.Sprintf("my/ships/%s/sell", s.Symbol), SellRequest{
		Symbol: cargoSymbol,
		Units:  units,
	})
	if sellResult != nil {
		s.Cargo = sellResult.Cargo
	}
	return sellResult, err
}

func (s *Ship) Extract(ctx context.Context) (*ExtractionResult, *http.HttpError) {
	extractionResult, err := http.Request[ExtractionResult](ctx, "POST", fmt.Sprintf("my/ships/%s/extract", s.Symbol), nil)
	if extractionResult != nil {
		s.Cargo = extractionResult.Cargo
	}
	return extractionResult, err
}

func (s *Ship) ExtractSurvey(ctx context.Context, survey *Survey) (*ExtractionResult, *http.HttpError) {
	extractionResult, err := http.Request[ExtractionResult](ctx, "POST", fmt.Sprintf("my/ships/%s/extract", s.Symbol), map[string]*Survey{
		"survey": survey,
	})
	if extractionResult != nil {
		s.Cargo = extractionResult.Cargo
	}
	return extractionResult, err
}

func (s *Ship) Survey(ctx context.Context) (*SurveyResult, *http.HttpError) {
	return http.Request[SurveyResult](ctx, "POST", fmt.Sprintf("my/ships/%s/survey", s.Symbol), nil)
}

func (s *Ship) Jump(ctx context.Context, waypoint Waypoint) (*ShipJumpResult, *http.HttpError) {
	jumpResult, err := http.Request[ShipJumpResult](ctx, "POST", fmt.Sprintf("my/ships/%s/jump", s.Symbol), map[string]Waypoint{
		"waypointSymbol": waypoint,
	})
	if err == nil {
		s.Nav = jumpResult.Nav
		return jumpResult, nil
	}
	return jumpResult, err
}

func (s *Ship) Warp(ctx context.Context, waypoint Waypoint) (*ShipWarpResult, *http.HttpError) {
	if !s.CanWarp() {
		return nil, &http.HttpError{Code: http.ErrNoWarpDrive, Message: "No Warp drive"}
	}
	warpResult, err := http.Request[ShipWarpResult](ctx, "POST", fmt.Sprintf("my/ships/%s/warp", s.Symbol), map[string]Waypoint{
		"waypointSymbol": waypoint,
	})
	if err == nil {
		s.Fuel = warpResult.Fuel
		s.Nav = warpResult.Nav
	}
	return warpResult, err
}

func (s *Ship) Chart(ctx context.Context) (*ShipChartResult, *http.HttpError) {
	return http.Request[ShipChartResult](ctx, "POST", fmt.Sprintf("my/ships/%s/chart", s.Symbol), nil)
}

func (s *Ship) ScanWaypoints(ctx context.Context) (*ShipScanWaypointsResult, *http.HttpError) {
	return http.Request[ShipScanWaypointsResult](ctx, "POST", fmt.Sprintf("my/ships/%s/scan/waypoints", s.Symbol), nil)
}

func (s *Ship) ScanSystems(ctx context.Context) (*ShipScanSystemsResult, *http.HttpError) {
	return http.Request[ShipScanSystemsResult](ctx, "POST", fmt.Sprintf("my/ships/%s/scan/systems", s.Symbol), nil)
}

func (s *Ship) JettisonCargo(ctx context.Context, symbol string, amount int) *http.HttpError {
	shipUpdate, err := http.Request[Ship](ctx, "POST", fmt.Sprintf("my/ships/%s/jettison", s.Symbol), map[string]any{
		"symbol": symbol,
		"units":  amount,
	})

	if err != nil && shipUpdate != nil {
		fmt.Println(shipUpdate)
		s.Cargo = shipUpdate.Cargo
	}
	return err
}

func (s *Ship) TransferCargo(ctx context.Context, ship string, symbol string, amount int) *http.HttpError {
	shipUpdate, err := http.Request[Ship](ctx, "POST", fmt.Sprintf("my/ships/%s/transfer", s.Symbol), map[string]any{
		"tradeSymbol": symbol,
		"shipSymbol":  ship,
		"units":       amount,
	})

	if err == nil {
		s.Cargo = shipUpdate.Cargo
	}
	return err
}

func (s *Ship) GetCargo(ctx context.Context) (*ShipCargo, *http.HttpError) {
	cargo, err := http.Request[ShipCargo](ctx, "GET", fmt.Sprintf("my/ships/%s/cargo", s.Symbol), nil)

	if err == nil {
		s.Cargo = *cargo
	}

	return cargo, err
}

func (s *Ship) Purchase(ctx context.Context, symbol string, amount int) (*ItemPurchaseResult, *http.HttpError) {
	result, err := http.Request[ItemPurchaseResult](ctx, "POST", fmt.Sprintf("my/ships/%s/purchase", s.Symbol), map[string]any{
		"symbol": symbol,
		"units":  amount,
	})

	if err == nil {
		s.Cargo = *result.Cargo
	}

	return result, err
}

func (s *Ship) NegotiateContract(ctx context.Context) (*Contract, *http.HttpError) {
	result, err := http.Request[map[string]*Contract](ctx, "POST", fmt.Sprintf("my/ships/%s/negotiate/contract", s.Symbol), nil)

	if err != nil {
		return nil, err
	}

	return (*result)["contract"], err
}

func (s *Ship) Refine(ctx context.Context, produce string) (*ShipRefineResult, *http.HttpError) {
	result, err := http.Request[ShipRefineResult](ctx, "POST", fmt.Sprintf("my/ships/%s/refine", s.Symbol), map[string]string{
		"produce": produce,
	})

	if err == nil {
		s.Cargo = result.Cargo
	}

	return result, err
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
	Departure     WaypointData `json:"origin"`
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

func (sf *ShipFuel) IsFull() bool {
	return sf.Current >= sf.Capacity
}

func (sf *ShipFuel) GetRemainingCapacity() int {
	return sf.Capacity - sf.Current
}

type ShipFrame struct {
	Symbol         string          `json:"symbol"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	ModuleSlots    int             `json:"moduleSlots"`
	MountingPoints int             `json:"mountingPoints"`
	FuelCapacity   int             `json:"fuelCapacity"`
	Condition      float64         `json:"condition"`
	Integrity      float64         `json:"integrity"`
	Quality        int             `json:"quality"`
	Requirements   ShipRequirement `json:"requirements"`
}

type ShipReactor struct {
	Symbol       string          `json:"symbol"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Condition    float64         `json:"condition"`
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

func (sc *ShipCargo) IsFull() bool {
	return sc.Units >= sc.Capacity
}

func (sc *ShipCargo) GetRemainingCapacity() int {
	return sc.Capacity - sc.Units
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
