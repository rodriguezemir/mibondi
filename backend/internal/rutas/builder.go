package rutas

import (
	"sort"
	"strings"

	"mibondi.github.com/internal/gtfs"
)

func Build(routes []gtfs.Route, trips []gtfs.Trip) []Ruta {
	routeByID := make(map[string]gtfs.Route, len(routes))
	for _, r := range routes {
		routeByID[r.ID] = r
	}

	seen := make(map[string]struct{}, len(trips))
	out := make([]Ruta, 0, len(trips))
	for _, t := range trips {
		r, ok := routeByID[t.RouteID]
		if !ok {
			continue
		}
		lineaID := gtfs.DisplayName(r)
		key := lineaID + "|" + t.Headsign
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, Ruta{
			ID:        slugID(lineaID, t.Headsign),
			LineaID:   lineaID,
			RouteID:   t.RouteID,
			Headsign:  t.Headsign,
			Direction: t.Direction,
			ShapeID:   t.ShapeID,
			AgencyID:  r.AgencyID,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].LineaID != out[j].LineaID {
			return out[i].LineaID < out[j].LineaID
		}
		return out[i].Headsign < out[j].Headsign
	})

	return out
}

func slugID(lineaID, headsign string) string {
	return lineaID + "-" + slugify(headsign)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := true
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}
