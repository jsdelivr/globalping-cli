package cmd

import (
	"strings"

	"github.com/jsdelivr/globalping-cli/model"
)

func createLocations(from string) []model.Locations {
	fromArr := strings.Split(from, ",")
	locations := make([]model.Locations, len(fromArr))
	for i, v := range fromArr {
		locations[i] = model.Locations{
			Magic: strings.TrimSpace(v),
		}
	}
	return locations
}

func inProgressUpdates(ci bool) bool {
	return !(ci)
}
