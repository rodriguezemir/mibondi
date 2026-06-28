package gtfs

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func Load(dir string) (*Feed, error) {
	agencies, err := loadAgencies(filepath.Join(dir, "agency.txt"))
	if err != nil {
		return nil, fmt.Errorf("agencies: %w", err)
	}

	routes, err := loadRoutes(filepath.Join(dir, "routes.txt"))
	if err != nil {
		return nil, fmt.Errorf("routes: %w", err)
	}

	trips, err := loadTrips(filepath.Join(dir, "trips.txt"))
	if err != nil {
		return nil, fmt.Errorf("trips: %w", err)
	}

	shapePoints, err := loadShapes(filepath.Join(dir, "shapes.txt"))
	if err != nil {
		return nil, fmt.Errorf("shapes: %w", err)
	}

	stops, err := loadStops(filepath.Join(dir, "stops.txt"))
	if err != nil {
		return nil, fmt.Errorf("stops: %w", err)
	}

	feed := &Feed{
		Agencies:    agencies,
		Lineas:      buildLineas(routes),
		Routes:      routes,
		Trips:       trips,
		ShapePoints: shapePoints,
		Stops:       stops,
	}

	sort.Slice(feed.Agencies, func(i, j int) bool {
		return feed.Agencies[i].ID < feed.Agencies[j].ID
	})
	sort.Slice(feed.Lineas, func(i, j int) bool {
		if feed.Lineas[i].AgencyID != feed.Lineas[j].AgencyID {
			return feed.Lineas[i].AgencyID < feed.Lineas[j].AgencyID
		}
		return feed.Lineas[i].ID < feed.Lineas[j].ID
	})

	return feed, nil
}

func loadAgencies(path string) ([]Agency, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	idx := indexColumns(rows[0])
	out := make([]Agency, 0, len(rows)-1)
	for _, row := range rows[1:] {
		out = append(out, Agency{
			ID:       row[idx["agency_id"]],
			Name:     row[idx["agency_name"]],
			URL:      row[idx["agency_url"]],
			Timezone: row[idx["agency_timezone"]],
			Lang:     row[idx["agency_lang"]],
			Phone:    row[idx["agency_phone"]],
		})
	}
	return out, nil
}

func loadRoutes(path string) ([]Route, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	idx := indexColumns(rows[0])
	out := make([]Route, 0, len(rows)-1)
	for _, row := range rows[1:] {
		out = append(out, Route{
			ID:        row[idx["route_id"]],
			AgencyID:  row[idx["agency_id"]],
			ShortName: row[idx["route_short_name"]],
			LongName:  row[idx["route_long_name"]],
			Color:     row[idx["route_color"]],
			TextColor: row[idx["route_text_color"]],
		})
	}
	return out, nil
}

func loadTrips(path string) ([]Trip, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	idx := indexColumns(rows[0])
	out := make([]Trip, 0, len(rows)-1)
	for _, row := range rows[1:] {
		dir, _ := strconv.Atoi(row[idx["direction_id"]])
		out = append(out, Trip{
			TripID:    row[idx["trip_id"]],
			RouteID:   row[idx["route_id"]],
			ServiceID: row[idx["service_id"]],
			Headsign:  row[idx["trip_headsign"]],
			Direction: dir,
			ShapeID:   row[idx["shape_id"]],
		})
	}
	return out, nil
}

func buildLineas(routes []Route) []Linea {
	seen := make(map[string]struct{}, len(routes))
	out := make([]Linea, 0, len(routes))
	for _, r := range routes {
		name := DisplayName(r)
		key := r.AgencyID + "|" + name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, Linea{
			ID:        name,
			ShortName: r.ShortName,
			LongName:  r.LongName,
			AgencyID:  r.AgencyID,
			Color:     r.Color,
			TextColor: r.TextColor,
		})
	}
	return out
}

func loadStopTimes(path string) ([]StopTime, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	idx := indexColumns(rows[0])
	out := make([]StopTime, 0, len(rows)-1)
	for _, row := range rows[1:] {
		seq, _ := strconv.Atoi(row[idx["stop_sequence"]])
		out = append(out, StopTime{
			TripID:        row[idx["trip_id"]],
			StopID:        row[idx["stop_id"]],
			ArrivalTime:   row[idx["arrival_time"]],
			DepartureTime: row[idx["departure_time"]],
			StopSequence:  seq,
		})
	}
	return out, nil
}

func LoadStopTimes(dir string) ([]StopTime, error) {
	return loadStopTimes(filepath.Join(dir, "stop_times.txt"))
}

func loadStops(path string) ([]Stop, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	idx := indexColumns(rows[0])
	out := make([]Stop, 0, len(rows)-1)
	for _, row := range rows[1:] {
		lat, _ := strconv.ParseFloat(row[idx["stop_lat"]], 64)
		lon, _ := strconv.ParseFloat(row[idx["stop_lon"]], 64)
		out = append(out, Stop{
			StopID:   row[idx["stop_id"]],
			StopName: row[idx["stop_name"]],
			Lat:      lat,
			Lon:      lon,
		})
	}
	return out, nil
}

func LoadStops(dir string) ([]Stop, error) {
	return loadStops(filepath.Join(dir, "stops.txt"))
}

func BuildShapeStops(trips []Trip, stops []Stop, stopTimes []StopTime) map[string][]Stop {
	tripToShape := make(map[string]string, len(trips))
	for _, t := range trips {
		tripToShape[t.TripID] = t.ShapeID
	}

	stopByID := make(map[string]Stop, len(stops))
	for _, s := range stops {
		stopByID[s.StopID] = s
	}

	setByShape := make(map[string]map[string]struct{})
	for _, st := range stopTimes {
		shapeID, ok := tripToShape[st.TripID]
		if !ok {
			continue
		}
		if _, ok := setByShape[shapeID]; !ok {
			setByShape[shapeID] = make(map[string]struct{})
		}
		setByShape[shapeID][st.StopID] = struct{}{}
	}

	result := make(map[string][]Stop, len(setByShape))
	for shapeID, set := range setByShape {
		out := make([]Stop, 0, len(set))
		for stopID := range set {
			if s, ok := stopByID[stopID]; ok {
				out = append(out, s)
			}
		}
		result[shapeID] = out
	}
	return result
}

func loadShapes(path string) ([]ShapePoint, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}
	idx := indexColumns(rows[0])
	out := make([]ShapePoint, 0, len(rows)-1)
	for _, row := range rows[1:] {
		lat, _ := strconv.ParseFloat(row[idx["shape_pt_lat"]], 64)
		lon, _ := strconv.ParseFloat(row[idx["shape_pt_lon"]], 64)
		seq, _ := strconv.Atoi(row[idx["shape_pt_sequence"]])
		out = append(out, ShapePoint{
			ShapeID:  row[idx["shape_id"]],
			Lat:      lat,
			Lon:      lon,
			Sequence: seq,
		})
	}
	return out, nil
}

func DisplayName(r Route) string {
	if s := strings.TrimSpace(r.ShortName); s != "" {
		return s
	}
	if s := strings.TrimSpace(r.LongName); s != "" {
		return s
	}
	return r.ID
}

func readCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	return rows, nil
}

func indexColumns(header []string) map[string]int {
	out := make(map[string]int, len(header))
	for i, col := range header {
		out[strings.TrimSpace(col)] = i
	}
	return out
}
