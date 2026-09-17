package supabase

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ponytojas/go-mqtt-timescale/config"
	"github.com/ponytojas/go-mqtt-timescale/internal/models"
)

// TestSupabaseInsertAndDelete verifies the configured Supabase API credentials,
// schema exposure, and RLS policies. It is opt-in because it writes real data.
func TestSupabaseInsertAndDelete(t *testing.T) {
	if os.Getenv("SUPABASE_INTEGRATION_TEST") != "true" {
		t.Skip("set SUPABASE_INTEGRATION_TEST=true to run against Supabase")
	}

	cfg := &config.Config{Supabase: config.SupabaseConfig{
		URL:       os.Getenv("SUPABASE_URL"),
		APIKey:    os.Getenv("SUPABASE_API_KEY"),
		Schema:    envOrDefault("SUPABASE_SCHEMA", "labtools"),
		TableName: envOrDefault("SUPABASE_TABLE_NAME", "sensor_data"),
	}}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	deviceID := os.Getenv("SUPABASE_TEST_DEVICE_ID")
	if deviceID == "" {
		t.Fatal("SUPABASE_TEST_DEVICE_ID is required and must match an existing sensor ID")
	}

	// PostgreSQL timestamps have microsecond precision. Matching that precision
	// lets cleanup target precisely the row this test created.
	timestamp := time.Now().UTC().Truncate(time.Microsecond)
	data := &models.SensorData{
		Timestamp:   timestamp,
		Temperature: 20.5,
		Humidity:    45.0,
		Light:       100.0,
		DeviceID:    deviceID,
	}

	if err := client.InsertSensorData(context.Background(), data); err != nil {
		t.Fatalf("insert test row: %v", err)
	}
	t.Cleanup(func() {
		_, _, err := client.client.From(client.tableName).
			Delete("minimal", "").
			Eq("time", timestamp.Format(time.RFC3339Nano)).
			Eq("device_id", deviceID).
			Execute()
		if err != nil {
			t.Errorf("remove test row for %s: %v", deviceID, err)
		}
	})
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
