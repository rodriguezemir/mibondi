package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"mibondi.github.com/internal/gtfs"
	"mibondi.github.com/internal/rutas"
	"mibondi.github.com/internal/shapes"
)

const dataDir = "gfts"

func main() {
	feed, err := gtfs.Load(dataDir)
	if err != nil {
		log.Fatalf("loading GTFS data: %v", err)
	}
	log.Printf("loaded %d agencies, %d lineas, %d routes, %d trips, %d shape points",
		len(feed.Agencies), len(feed.Lineas), len(feed.Routes), len(feed.Trips), len(feed.ShapePoints))

	rutasList := rutas.Build(feed.Routes, feed.Trips)
	log.Printf("built %d rutas", len(rutasList))

	shapesList := shapes.Build(feed.ShapePoints)
	log.Printf("built %d shapes", len(shapesList))

	rutasHandler := rutas.New(rutasList)
	shapesHandler := shapes.New(shapesList)

	r := gin.Default()
	r.Use(cors.Default())

	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		api.GET("/agencias", listAgencias(feed))
		api.GET("/lineas", listLineas(feed))
		api.GET("/rutas", rutasHandler.List)
		api.GET("/rutas/:short_name", rutasHandler.GetByLinea)
		api.GET("/shapes/:shape_id", shapesHandler.Get)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func listAgencias(feed *gtfs.Feed) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, feed.Agencies)
	}
}

func listLineas(feed *gtfs.Feed) gin.HandlerFunc {
	return func(c *gin.Context) {
		agencyID := c.Query("agencia")
		if agencyID == "" {
			c.JSON(http.StatusOK, feed.Lineas)
			return
		}
		filtered := make([]gtfs.Linea, 0, len(feed.Lineas))
		for _, l := range feed.Lineas {
			if l.AgencyID == agencyID {
				filtered = append(filtered, l)
			}
		}
		c.JSON(http.StatusOK, filtered)
	}
}
