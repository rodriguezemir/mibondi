package shapes

type Shape struct {
	ID     string  `json:"shape_id"`
	Points []Point `json:"points"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
