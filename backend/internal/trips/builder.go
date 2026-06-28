package trips

import (
	"sort"

	"mibondi.github.com/internal/gtfs"
)

func Build(rawTrips []gtfs.Trip) []Trip {
	out := make([]Trip, len(rawTrips))
	for i, t := range rawTrips {
		out[i] = Trip{
			TripID:    t.TripID,
			RouteID:   t.RouteID,
			ServiceID: t.ServiceID,
			Headsign:  t.Headsign,
			Direction: t.Direction,
			ShapeID:   t.ShapeID,
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].TripID < out[j].TripID
	})
	return out
}
