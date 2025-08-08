package routine

import (
	"github.com/stretchr/testify/assert"
	"spacetraders/entity"
	"testing"
)

func TestGoToMiningArea_IsWaypointBlacklisted(t *testing.T) {
	t.Run("Blacklist present", func(t *testing.T) {
		g := GoToMiningArea{
			blacklist: []entity.Waypoint{"WAYPOINT-1"},
		}
		assert.True(t, g.IsWaypointBlacklisted("WAYPOINT-1"))
		assert.False(t, g.IsWaypointBlacklisted("WAYPOINT-2"))
	})

	t.Run("Blacklist nil", func(t *testing.T) {
		g := GoToMiningArea{}
		assert.False(t, g.IsWaypointBlacklisted("WAYPOINT-1"))
		assert.False(t, g.IsWaypointBlacklisted("WAYPOINT-2"))
	})

	t.Run("Blacklist empty", func(t *testing.T) {
		g := GoToMiningArea{
			blacklist: []entity.Waypoint{},
		}
		assert.False(t, g.IsWaypointBlacklisted("WAYPOINT-1"))
		assert.False(t, g.IsWaypointBlacklisted("WAYPOINT-2"))
	})
}

func TestGoToMiningArea_ScoreWaypoint(t *testing.T) {
	t.Run("Ineligible Waypoints", func(t *testing.T) {

	})
}
