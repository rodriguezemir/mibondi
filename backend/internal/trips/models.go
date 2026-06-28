package trips

type Trip struct {
	TripID    string `json:"trip_id"`
	RouteID   string `json:"route_id"`
	ServiceID string `json:"service_id"`
	Headsign  string `json:"headsign"`
	Direction int    `json:"direction"`
	ShapeID   string `json:"shape_id"`
}
