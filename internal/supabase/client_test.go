package supabase

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ponytojas/go-mqtt-timescale/config"
	"github.com/ponytojas/go-mqtt-timescale/internal/models"
)

func TestInsertSensorDataUsesSupabaseDataAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rest/v1/sensor_data" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("apikey") != "secret" || r.Header.Get("Content-Profile") != "labtools" {
			t.Errorf("unexpected Supabase headers: %#v", r.Header)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"time"`) || !strings.Contains(string(body), `"device_id":"sensor-1"`) {
			t.Errorf("unexpected body: %s", body)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client, err := NewClient(&config.Config{Supabase: config.SupabaseConfig{
		URL: server.URL, APIKey: "secret", Schema: "labtools", TableName: "sensor_data",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.InsertSensorData(context.Background(), &models.SensorData{Timestamp: time.Now(), DeviceID: "sensor-1"}); err != nil {
		t.Fatal(err)
	}
}

func TestNewClientRequiresSupabaseCredentials(t *testing.T) {
	_, err := NewClient(&config.Config{Supabase: config.SupabaseConfig{Schema: "labtools", TableName: "sensor_data"}})
	if err == nil || !strings.Contains(err.Error(), "SUPABASE_URL") {
		t.Fatalf("error = %v, want missing URL error", err)
	}
}
