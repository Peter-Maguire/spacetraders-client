package constant

type WaypointType string

const (
	WaypointTypePlanet                WaypointType = "PLANET"
	WaypointTypeGasGiant              WaypointType = "GAS_GIANT"
	WaypointTypeMoon                  WaypointType = "MOON"
	WaypointTypeOrbitalStation        WaypointType = "ORBITAL_STATION"
	WaypointTypeJumpGate              WaypointType = "JUMP_GATE"
	WaypointTypeAsteroidField         WaypointType = "ASTEROID_FIELD"
	WaypointTypeAsteroidBase          WaypointType = "ASTEROID_BASE"
	WaypointTypeNebula                WaypointType = "NEBULA"
	WaypointTypeDebrisField           WaypointType = "DEBRIS_FIELD"
	WaypointTypeGravityWell           WaypointType = "GRAVITY_WELL"
	WaypointTypeArtificialGravityWell WaypointType = "ARTIFICIAL_GRAVITY_WELL"
	WaypointTypeFuelStation           WaypointType = "FUEL_STATION"
)
