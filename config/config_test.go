package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestGetMQTTBrokerURL(t *testing.T) {
	tests := []struct {
		broker, expected string
		port             int
	}{
		{"https://mqtt.example.com", "ssl://mqtt.example.com:8883", 8883},
		{"http://mqtt.example.com", "tcp://mqtt.example.com:1883", 1883},
		{"tcp://mqtt.example.com:1883", "tcp://mqtt.example.com:1883", 1883},
		{"mqtt.example.com", "tcp://mqtt.example.com:1883", 1883},
	}
	for _, test := range tests {
		cfg := &Config{MQTT: MQTTConfig{Broker: test.broker, Port: test.port}}
		if got := cfg.GetMQTTBrokerURL(); got != test.expected {
			t.Errorf("GetMQTTBrokerURL() = %q, want %q", got, test.expected)
		}
	}
}

func TestLoadSupabaseConfigFromEnvironment(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("SUPABASE_URL", "https://project-ref.supabase.co")
	t.Setenv("SUPABASE_API_KEY", "sb_secret_example")
	t.Setenv("SUPABASE_SCHEMA", "labtools")
	t.Setenv("SUPABASE_TABLE_NAME", "sensor_data")

	cfg, err := LoadConfig("./does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Supabase.URL != "https://project-ref.supabase.co" || cfg.Supabase.APIKey != "sb_secret_example" {
		t.Error("Supabase URL or API key was not loaded")
	}
	if cfg.Supabase.Schema != "labtools" || cfg.Supabase.TableName != "sensor_data" {
		t.Errorf("target = %s.%s, want labtools.sensor_data", cfg.Supabase.Schema, cfg.Supabase.TableName)
	}
}

func TestDefaultSupabaseTarget(t *testing.T) {
	cfg := GetDefaultConfig()
	if cfg.Supabase.Schema != "labtools" || cfg.Supabase.TableName != "sensor_data" {
		t.Errorf("target = %s.%s, want labtools.sensor_data", cfg.Supabase.Schema, cfg.Supabase.TableName)
	}
}
