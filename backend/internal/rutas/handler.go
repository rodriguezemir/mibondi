package rutas

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	rutas []Ruta
}

func New(rutas []Ruta) *Handler {
	return &Handler{rutas: rutas}
}

func (h *Handler) List(c *gin.Context) {
	direction, err := parseDirection(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filtered := h.rutas
	if agencyID := c.Query("agencia"); agencyID != "" {
		filtered = filter(filtered, func(r Ruta) bool {
			return r.AgencyID == agencyID
		})
	}
	if lineaID := c.Query("linea"); lineaID != "" {
		filtered = filter(filtered, func(r Ruta) bool {
			return r.LineaID == lineaID
		})
	}
	if direction != nil {
		d := *direction
		filtered = filter(filtered, func(r Ruta) bool {
			return r.Direction == d
		})
	}
	c.JSON(http.StatusOK, filtered)
}

func (h *Handler) GetByLinea(c *gin.Context) {
	shortName := c.Param("short_name")

	direction, err := parseDirection(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filtered := h.rutas
	if agencyID := c.Query("agencia"); agencyID != "" {
		filtered = filter(filtered, func(r Ruta) bool {
			return r.AgencyID == agencyID
		})
	}
	if direction != nil {
		d := *direction
		filtered = filter(filtered, func(r Ruta) bool {
			return r.Direction == d
		})
	}
	filtered = filter(filtered, func(r Ruta) bool {
		return r.LineaID == shortName
	})

	if len(filtered) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "no rutas found",
			"short_name": shortName,
		})
		return
	}
	c.JSON(http.StatusOK, filtered)
}

func parseDirection(c *gin.Context) (*int, error) {
	s := c.Query("direction")
	if s == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("direction must be a number, got %q", s)
	}
	if n != 0 && n != 1 {
		return nil, fmt.Errorf("direction must be 0 or 1, got %d", n)
	}
	return &n, nil
}

func filter(rutas []Ruta, keep func(Ruta) bool) []Ruta {
	out := make([]Ruta, 0, len(rutas))
	for _, r := range rutas {
		if keep(r) {
			out = append(out, r)
		}
	}
	return out
}
