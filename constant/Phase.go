package constant

type Phase string

const (
	// PhaseExplore is when we have not fully explored all waypoints and shipyards, we focus on visiting each one
	PhaseExplore Phase = "EXPLORE"

	// PhaseExpandFleet is when we need to build up a fleet of mining drones to fulfill contracts (when does this end?)
	PhaseExpandFleet Phase = "EXPAND_FLEET"

	// PhaseBuildJumpGate is when we stop focussing on contracts and start building the jump gate
	PhaseBuildJumpGate Phase = "BUILD_JUMP_GATE"

	// PhaseChartSystems is when we start charting uncharted systems
	PhaseChartSystems Phase = "CHART_SYSTEMS"
)
