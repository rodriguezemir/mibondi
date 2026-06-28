package shapes

import (
	"sort"

	"mibondi.github.com/internal/gtfs"
)

func Build(points []gtfs.ShapePoint) []Shape {
	byID := make(map[string][]gtfs.ShapePoint, len(points)/4)
	for _, p := range points {
		byID[p.ShapeID] = append(byID[p.ShapeID], p)
	}

	out := make([]Shape, 0, len(byID))
	for id, pts := range byID {
		sort.Slice(pts, func(i, j int) bool {
			return pts[i].Sequence < pts[j].Sequence
		})
		converted := make([]Point, len(pts))
		for i, p := range pts {
			converted[i] = Point{Lat: p.Lat, Lon: p.Lon}
		}
		out = append(out, Shape{ID: id, Points: converted})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	return out
}
