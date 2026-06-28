package database

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mibondi.github.com/internal/gtfs"
)

type Shape struct {
	ShapeID  string  `gorm:"column:shape_id;type:text;not null;index"`
	Lat      float64 `gorm:"column:lat;type:double precision;not null"`
	Lon      float64 `gorm:"column:lon;type:double precision;not null"`
	Sequence int     `gorm:"column:sequence;type:integer;not null"`
}

func (Shape) TableName() string { return "shapes" }

type StopTime struct {
	TripID        string `gorm:"column:trip_id;type:text;not null;index"`
	StopID        string `gorm:"column:stop_id;type:text;not null;index"`
	ArrivalTime   string `gorm:"column:arrival_time;type:text;not null"`
	DepartureTime string `gorm:"column:departure_time;type:text;not null"`
	StopSequence  int    `gorm:"column:stop_sequence;type:integer;not null"`
}

func (StopTime) TableName() string { return "stop_times" }

type Stop struct {
	StopID   string  `gorm:"column:stop_id;type:text;not null;index"`
	StopName string  `gorm:"column:stop_name;type:text;not null"`
	Lat      float64 `gorm:"column:stop_lat;type:double precision;not null"`
	Lon      float64 `gorm:"column:stop_lon;type:double precision;not null"`
}

func (Stop) TableName() string { return "stops" }

type Trip struct {
	TripID    string `gorm:"column:trip_id;type:text;not null;index"`
	RouteID   string `gorm:"column:route_id;type:text;not null;index"`
	ServiceID string `gorm:"column:service_id;type:text;not null;index"`
	Headsign  string `gorm:"column:trip_headsign;type:text;not null"`
	Direction int    `gorm:"column:direction_id;type:integer;not null"`
	ShapeID   string `gorm:"column:shape_id;type:text;not null;index"`
}

func (Trip) TableName() string { return "trips" }

type DB struct {
	gorm *gorm.DB
}

func New(ctx context.Context, dsn string) (*DB, error) {
	g, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := g.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying *sql.DB: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{gorm: g}, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.gorm.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) ImportGTFS(ctx context.Context, feed *gtfs.Feed, stopTimes []gtfs.StopTime) error {
	if err := db.migrate(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := db.importShapes(ctx, feed.ShapePoints); err != nil {
		return fmt.Errorf("import shapes: %w", err)
	}
	if err := db.importStops(ctx, feed.Stops); err != nil {
		return fmt.Errorf("import stops: %w", err)
	}
	if err := db.importTrips(ctx, feed.Trips); err != nil {
		return fmt.Errorf("import trips: %w", err)
	}
	if err := db.importStopTimes(ctx, stopTimes); err != nil {
		return fmt.Errorf("import stop_times: %w", err)
	}
	return nil
}

func (db *DB) migrate(ctx context.Context) error {
	return db.gorm.WithContext(ctx).AutoMigrate(&Shape{}, &StopTime{}, &Stop{}, &Trip{})
}

func (db *DB) importShapes(ctx context.Context, points []gtfs.ShapePoint) error {
	if count, err := db.count(ctx, &Shape{}); err != nil {
		return err
	} else if count > 0 {
		log.Printf("database: shapes already has %d rows, skipping", count)
		return nil
	}
	if len(points) == 0 {
		return nil
	}

	rows := make([]Shape, len(points))
	for i, p := range points {
		rows[i] = Shape{
			ShapeID:  p.ShapeID,
			Lat:      p.Lat,
			Lon:      p.Lon,
			Sequence: p.Sequence,
		}
	}
	log.Printf("database: importing %d shape points", len(rows))
	return db.gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(rows, 1000).Error
	})
}

func (db *DB) importStops(ctx context.Context, stops []gtfs.Stop) error {
	if count, err := db.count(ctx, &Stop{}); err != nil {
		return err
	} else if count > 0 {
		log.Printf("database: stops already has %d rows, skipping", count)
		return nil
	}
	if len(stops) == 0 {
		return nil
	}

	rows := make([]Stop, len(stops))
	for i, s := range stops {
		rows[i] = Stop{
			StopID:   s.StopID,
			StopName: s.StopName,
			Lat:      s.Lat,
			Lon:      s.Lon,
		}
	}
	log.Printf("database: importing %d stops", len(rows))
	return db.gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(rows, 1000).Error
	})
}

func (db *DB) importTrips(ctx context.Context, rawTrips []gtfs.Trip) error {
	if count, err := db.count(ctx, &Trip{}); err != nil {
		return err
	} else if count > 0 {
		log.Printf("database: trips already has %d rows, skipping", count)
		return nil
	}
	if len(rawTrips) == 0 {
		return nil
	}

	rows := make([]Trip, len(rawTrips))
	for i, t := range rawTrips {
		rows[i] = Trip{
			TripID:    t.TripID,
			RouteID:   t.RouteID,
			ServiceID: t.ServiceID,
			Headsign:  t.Headsign,
			Direction: t.Direction,
			ShapeID:   t.ShapeID,
		}
	}
	log.Printf("database: importing %d trips", len(rows))
	return db.gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(rows, 1000).Error
	})
}

func (db *DB) importStopTimes(ctx context.Context, times []gtfs.StopTime) error {
	if count, err := db.count(ctx, &StopTime{}); err != nil {
		return err
	} else if count > 0 {
		log.Printf("database: stop_times already has %d rows, skipping", count)
		return nil
	}
	if len(times) == 0 {
		return nil
	}

	rows := make([]StopTime, len(times))
	for i, t := range times {
		rows[i] = StopTime{
			TripID:        t.TripID,
			StopID:        t.StopID,
			ArrivalTime:   t.ArrivalTime,
			DepartureTime: t.DepartureTime,
			StopSequence:  t.StopSequence,
		}
	}
	log.Printf("database: importing %d stop_times", len(rows))
	return db.gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(rows, 5000).Error
	})
}

func (db *DB) count(ctx context.Context, model any) (int64, error) {
	var n int64
	err := db.gorm.WithContext(ctx).Model(model).Count(&n).Error
	return n, err
}
