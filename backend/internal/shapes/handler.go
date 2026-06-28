package shapes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mibondi.github.com/internal/gtfs"
)

type Shape struct {
	ID     string  `json:"shape_id"`
	Points []Point `json:"points"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Handler struct {
	shapes       map[string]Shape
	stopsByShape map[string][]gtfs.Stop
}

func New(shapes []Shape, stopsByShape map[string][]gtfs.Stop) *Handler {
	m := make(map[string]Shape, len(shapes))
	for _, s := range shapes {
		m[s.ID] = s
	}
	return &Handler{shapes: m, stopsByShape: stopsByShape}
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("shape_id")
	s, ok := h.shapes[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error":    "shape not found",
			"shape_id": id,
		})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) GetStops(c *gin.Context) {
	id := c.Param("shape_id")
	stops, ok := h.stopsByShape[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error":    "no stops for this shape",
			"shape_id": id,
		})
		return
	}
	c.JSON(http.StatusOK, stops)
}
