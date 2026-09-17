package supabase

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	supabasego "github.com/supabase-community/supabase-go"

	"github.com/ponytojas/go-mqtt-timescale/config"
	"github.com/ponytojas/go-mqtt-timescale/internal/models"
)

// Client stores readings through the community Supabase Go client.
type Client struct {
	client    *supabasego.Client
	tableName string
}

// NewClient configures the Supabase client for the project URL, API key, and schema.
func NewClient(cfg *config.Config) (*Client, error) {
	baseURL := strings.TrimRight(cfg.Supabase.URL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("invalid SUPABASE_URL: %w", err)
	}
	if cfg.Supabase.APIKey == "" {
		return nil, fmt.Errorf("SUPABASE_API_KEY is required")
	}
	if cfg.Supabase.Schema == "" || cfg.Supabase.TableName == "" {
		return nil, fmt.Errorf("SUPABASE_SCHEMA and SUPABASE_TABLE_NAME are required")
	}

	client, err := supabasego.NewClient(baseURL, cfg.Supabase.APIKey, &supabasego.ClientOptions{
		Schema: cfg.Supabase.Schema,
	})
	if err != nil {
		return nil, fmt.Errorf("create Supabase client: %w", err)
	}
	return &Client{client: client, tableName: cfg.Supabase.TableName}, nil
}

type sensorDataPayload struct {
	Time        time.Time `json:"time"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Light       float64   `json:"light"`
	DeviceID    string    `json:"device_id"`
}

// InsertSensorData inserts one row through Supabase's Data REST API.
func (c *Client) InsertSensorData(ctx context.Context, data *models.SensorData) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, _, err := c.client.From(c.tableName).Insert(sensorDataPayload{
		Time:        data.Timestamp,
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		Light:       data.Light,
		DeviceID:    data.DeviceID,
	}, false, "", "minimal", "").Execute()
	if err != nil {
		return fmt.Errorf("insert sensor data through Supabase: %w", err)
	}
	return nil
}
