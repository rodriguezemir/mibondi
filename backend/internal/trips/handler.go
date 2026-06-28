package trips

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	trips []Trip
}

func New(trips []Trip) *Handler {
	return &Handler{trips: trips}
}

func (h *Handler) List(c *gin.Context) {
	filtered := h.trips
	if shapeID := c.Query("shape_id"); shapeID != "" {
		filtered = filter(filtered, func(t Trip) bool {
			return t.ShapeID == shapeID
		})
	}
	if routeID := c.Query("route_id"); routeID != "" {
		filtered = filter(filtered, func(t Trip) bool {
			return t.RouteID == routeID
		})
	}
	if serviceID := c.Query("service_id"); serviceID != "" {
		filtered = filter(filtered, func(t Trip) bool {
			return t.ServiceID == serviceID
		})
	}
	c.JSON(http.StatusOK, filtered)
}

func (h *Handler) Get(c *gin.Context) {
	tripID := c.Param("trip_id")
	for _, t := range h.trips {
		if t.TripID == tripID {
			c.JSON(http.StatusOK, t)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error":   "trip not found",
		"trip_id": tripID,
	})
}

func filter(trips []Trip, keep func(Trip) bool) []Trip {
	out := make([]Trip, 0, len(trips))
	for _, t := range trips {
		if keep(t) {
			out = append(out, t)
		}
	}
	return out
}
