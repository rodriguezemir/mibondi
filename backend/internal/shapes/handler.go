package shapes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	shapes map[string]Shape
}

func New(shapes []Shape) *Handler {
	m := make(map[string]Shape, len(shapes))
	for _, s := range shapes {
		m[s.ID] = s
	}
	return &Handler{shapes: m}
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
