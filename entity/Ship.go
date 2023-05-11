package entity

import (
	"errors"
	"fmt"
	"spacetraders/http"
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

func (s *Ship) Navigate(waypoint Waypoint) (*ShipNav, *http.HttpError) {
	shipUpdate, err := http.Request[Ship]("POST", fmt.Sprintf("my/ships/%s/navigate", s.Symbol), NavigateRequest{
		WaypointSymbol: waypoint,
	})
	s.Nav = shipUpdate.Nav
	s.Fuel = shipUpdate.Fuel
	return &shipUpdate.Nav, err
}

func (s *Ship) Dock() error {
	shipNavUpdate, err := http.Request[Ship]("POST", fmt.Sprintf("my/ships/%s/dock", s.Symbol), nil)
	if shipNavUpdate != nil {
		s.Nav = shipNavUpdate.Nav
	}
	return err
}

func (s *Ship) Refuel() error {
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
