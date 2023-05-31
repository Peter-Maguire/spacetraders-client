package util

import "strings"

// Items other than ores that can be mined
var mineable = []string{"ICE_WATER", "QUARTZ_SAND", "AMMONIA_ICE", "SILICON_CRYSTALS"}

func IsMineable(item string) bool {
	if strings.HasSuffix(item, "_ORE") {
		return true
	}

	for _, m := range mineable {
		if m == item {
			return true
		}
	}
	return false
}

func IsRefineable(item string) bool {
	return strings.HasSuffix(item, "_ORE")
}
