package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"mibondi.github.com/internal/database"
	"mibondi.github.com/internal/gtfs"
	"mibondi.github.com/internal/rutas"
	"mibondi.github.com/internal/shapes"
	"mibondi.github.com/internal/trips"
)

const dataDir = "gfts"

func main() {
	ctx := context.Background()

	feed, err := gtfs.Load(dataDir)
	if err != nil {
		log.Fatalf("loading GTFS data: %v", err)
	}
	log.Printf("loaded %d agencies, %d lineas, %d routes, %d trips, %d shape points, %d stops",
		len(feed.Agencies), len(feed.Lineas), len(feed.Routes), len(feed.Trips), len(feed.ShapePoints), len(feed.Stops))

	stopTimes, err := gtfs.LoadStopTimes(dataDir)
	if err != nil {
		log.Fatalf("loading stop_times: %v", err)
	}
	log.Printf("loaded %d stop_times", len(stopTimes))

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		log.Printf("DATABASE_URL detected, importing GTFS to PostgreSQL...")
		db, err := database.New(ctx, dbURL)
		if err != nil {
			log.Fatalf("connect to database: %v", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("warn: close database: %v", err)
			}
		}()

		if err := db.ImportGTFS(ctx, feed, stopTimes); err != nil {
			log.Fatalf("import GTFS to database: %v", err)
		}
		log.Printf("database import complete")
	}

	rutasList := rutas.Build(feed.Routes, feed.Trips)
	log.Printf("built %d rutas", len(rutasList))

	shapesList := shapes.Build(feed.ShapePoints)
	log.Printf("built %d shapes", len(shapesList))

	tripsList := trips.Build(feed.Trips)
	log.Printf("built %d trips", len(tripsList))

	shapeStops := gtfs.BuildShapeStops(feed.Trips, feed.Stops, stopTimes)
	log.Printf("built stops index for %d shapes", len(shapeStops))

	rutasHandler := rutas.New(rutasList)
	shapesHandler := shapes.New(shapesList, shapeStops)
	tripsHandler := trips.New(tripsList)

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
		api.GET("/shapes/:shape_id/stops", shapesHandler.GetStops)
		api.GET("/trips", tripsHandler.List)
		api.GET("/trips/:trip_id", tripsHandler.Get)
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
