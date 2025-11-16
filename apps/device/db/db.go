package db

import (
	"context"
	"device/models"
	"fmt"

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

func (db *DB) CreateDevice(ctx context.Context, d models.Device) (models.Device, error) {
	query := `
		INSERT INTO devices (name, device_type_id, location)
		VALUES ($1, $2, $3)
		RETURNING id, device_type_id, name, location, status, connection_status, updated_at, created_at
	`

	var device models.Device
	err := db.Pool.QueryRow(ctx, query,
		d.Name,
		d.DeviceTypeID,
		d.Location,
	).Scan(
		&device.ID,
		&device.DeviceTypeID,
		&device.Name,
		&device.Location,
		&device.Status,
		&device.ConnectionStatus,
		&device.UpdatedAt,
		&device.CreatedAt,
	)
	if err != nil {
		return models.Device{}, fmt.Errorf("error creating device: %w", err)
	}

	return device, nil
}

func (db *DB) GetDevices(ctx context.Context) ([]models.Device, error) {
	query := `
		SELECT id, device_type_id, name, location, status, connection_status, updated_at, created_at
		FROM devices
		ORDER BY id
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying sensors: %w", err)
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var device models.Device
		err = rows.Scan(
			&device.ID,
			&device.DeviceTypeID,
			&device.Name,
			&device.Location,
			&device.Status,
			&device.ConnectionStatus,
			&device.UpdatedAt,
			&device.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning device row: %w", err)
		}
		devices = append(devices, device)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}

	return devices, nil
}
