package rutas

type Ruta struct {
	ID        string `json:"ruta_id"`
	LineaID   string `json:"linea_id"`
	RouteID   string `json:"route_id"`
	Headsign  string `json:"headsign"`
	Direction int    `json:"direction"`
	ShapeID   string `json:"shape_id"`
	AgencyID  string `json:"agency_id"`
}
