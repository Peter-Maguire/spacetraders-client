package util

import "strings"

// TODO deprecate in favour of Item
// Items other than ores that can be mined
var mineable = []string{"ICE_WATER", "QUARTZ_SAND", "AMMONIA_ICE", "SILICON_CRYSTALS"}
var siphonable = []string{"HYDROCARBON", "LIQUID_NITROGEN"}

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

func IsSiphonable(item string) bool {
	for _, m := range siphonable {
		if m == item {
			return true
		}
	}
	return false
}

func IsRefineable(item string) bool {
	return strings.HasSuffix(item, "_ORE")
}
