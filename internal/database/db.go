package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/ponytojas/go-mqtt-timescale/config"
	"github.com/ponytojas/go-mqtt-timescale/internal/models"
)

// TimescaleDB handles database operations
type TimescaleDB struct {
	conn   *pgx.Conn
	config *config.Config
}

// NewTimescaleDB creates a new TimescaleDB instance
func NewTimescaleDB(cfg *config.Config) (*TimescaleDB, error) {
	conn, err := pgx.Connect(context.Background(), cfg.GetDBConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &TimescaleDB{
		conn:   conn,
		config: cfg,
	}, nil
}

// Close closes the database connection
func (db *TimescaleDB) Close() error {
	return db.conn.Close(context.Background())
}

// InitializeTable checks if the table exists and creates it if it doesn't
func (db *TimescaleDB) InitializeTable() error {
	ctx := context.Background()
	tableName := db.config.Timescale.TableName

	// Check if table exists
	var exists bool
	err := db.conn.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		)
	`, tableName).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}

	// If table doesn't exist, create it
	if !exists {
		log.Printf("Creating table %s...", tableName)
		_, err = db.conn.Exec(ctx, fmt.Sprintf(`
			CREATE TABLE %s (
				time TIMESTAMPTZ NOT NULL,
				temperature DOUBLE PRECISION,
				humidity DOUBLE PRECISION,
				light DOUBLE PRECISION
				device_id TEXT NOT NULL,
			)
		`, tableName))

		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}

		// Convert to hypertable
		_, err = db.conn.Exec(ctx, fmt.Sprintf(`
			SELECT create_hypertable('%s', 'time')
		`, tableName))

		if err != nil {
			return fmt.Errorf("failed to convert table to hypertable: %w", err)
		}

		log.Printf("Table %s created and converted to hypertable", tableName)
	} else {
		log.Printf("Table %s already exists", tableName)
	}

	return nil
}

// InsertSensorData inserts sensor data into the database
func (db *TimescaleDB) InsertSensorData(data *models.SensorData) error {
	ctx := context.Background()
	tableName := db.config.Timescale.TableName

	_, err := db.conn.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (time, temperature, humidity, light, device_id)
		VALUES ($1, $2, $3, $4, $5)
	`, tableName), data.Timestamp, data.Temperature, data.Humidity, data.Light, data.Device_ID)

	if err != nil {
		return fmt.Errorf("failed to insert sensor data: %w", err)
	}

	return nil
}
