package db

import (
	"context"
	"fmt"
	"telemetry/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents the database connection
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new DB instance
func New(connString string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

func (db *DB) AddTemperature(ctx context.Context, s models.TelemetryData) error {
	query := `
		INSERT INTO telemetry_data (time, device_id, telemetry_type_id, numeric_value)
		VALUES (now(), $1, $2, $3)
		RETURNING time, device_id, telemetry_type_id, numeric_value
	`

	var data models.TelemetryData

	err := db.Pool.QueryRow(ctx, query,
		s.DeviceID,
		s.TelemetryTypeID,
		s.NumericValue,
	).Scan(
		&data.Time,
		&data.DeviceID,
		&data.TelemetryTypeID,
		&data.NumericValue,
	)
	if err != nil {
		return fmt.Errorf("error creating sensor: %w", err)
	}

	return nil
}
