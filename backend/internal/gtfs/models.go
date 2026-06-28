package gtfs

type Agency struct {
	ID       string `json:"agency_id"`
	Name     string `json:"agency_name"`
	URL      string `json:"agency_url"`
	Timezone string `json:"agency_timezone"`
	Lang     string `json:"agency_lang"`
	Phone    string `json:"agency_phone"`
}

type Linea struct {
	ID        string `json:"linea_id"`
	ShortName string `json:"short_name"`
	LongName  string `json:"long_name"`
	AgencyID  string `json:"agency_id"`
	Color     string `json:"color"`
	TextColor string `json:"text_color"`
}

type Route struct {
	ID        string
	AgencyID  string
	ShortName string
	LongName  string
	Color     string
	TextColor string
}

type Trip struct {
	TripID    string
	RouteID   string
	ServiceID string
	Headsign  string
	Direction int
	ShapeID   string
}

type ShapePoint struct {
	ShapeID  string
	Lat      float64
	Lon      float64
	Sequence int
}

type Stop struct {
	StopID   string  `json:"stop_id"`
	StopName string  `json:"stop_name"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
}

type StopTime struct {
	TripID        string
	StopID        string
	ArrivalTime   string
	DepartureTime string
	StopSequence  int
}

type Feed struct {
	Agencies    []Agency
	Lineas      []Linea
	Routes      []Route
	Trips       []Trip
	ShapePoints []ShapePoint
	Stops       []Stop
}
