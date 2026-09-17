package mqtt

import (
	"strings"
	"testing"

	"github.com/ponytojas/go-mqtt-timescale/config"
	"github.com/ponytojas/go-mqtt-timescale/internal/supabase"
)

func TestNewClientWithHTTPSURL(t *testing.T) {
	// Create a config with HTTPS URL
	cfg := &config.Config{
		MQTT: config.MQTTConfig{
			Broker:   "https://mqtt.ponytojas.dev",
			Port:     8883,
			ClientID: "test-client",
			Username: "test-user",
			Password: "test-password",
		},
	}

	// Get the broker URL
	brokerURL := cfg.GetMQTTBrokerURL()

	// Verify it's converted to SSL protocol
	if !strings.HasPrefix(brokerURL, "ssl://") {
		t.Errorf("Expected broker URL to start with ssl://, got %s", brokerURL)
	}

	// Verify the host is preserved
	if !strings.Contains(brokerURL, "mqtt.ponytojas.dev") {
		t.Errorf("Expected broker URL to contain mqtt.ponytojas.dev, got %s", brokerURL)
	}

	// Verify the port is included
	if !strings.Contains(brokerURL, ":8883") {
		t.Errorf("Expected broker URL to include port 8883, got %s", brokerURL)
	}

	// Test creating a client with this URL
	supabaseClient := &supabase.Client{}
	client, err := NewClient(cfg, supabaseClient)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}
}

func TestNewClientWithTCPURL(t *testing.T) {
	// Create a config with TCP URL
	cfg := &config.Config{
		MQTT: config.MQTTConfig{
			Broker:   "tcp://mqtt.example.com",
			Port:     1883,
			ClientID: "test-client",
		},
	}

	// Get the broker URL
	brokerURL := cfg.GetMQTTBrokerURL()

	// Verify it remains TCP protocol
	if !strings.HasPrefix(brokerURL, "tcp://") {
		t.Errorf("Expected broker URL to start with tcp://, got %s", brokerURL)
	}
}

func TestNewClientWithNoProtocol(t *testing.T) {
	// Create a config with no protocol
	cfg := &config.Config{
		MQTT: config.MQTTConfig{
			Broker:   "mqtt.example.com",
			Port:     1883,
			ClientID: "test-client",
		},
	}

	// Get the broker URL
	brokerURL := cfg.GetMQTTBrokerURL()

	// Verify it defaults to TCP protocol
	if !strings.HasPrefix(brokerURL, "tcp://") {
		t.Errorf("Expected broker URL to start with tcp://, got %s", brokerURL)
	}

	// Verify the port is included
	if !strings.Contains(brokerURL, ":1883") {
		t.Errorf("Expected broker URL to include port 1883, got %s", brokerURL)
	}
}

func TestNewClientWithHTTPURL(t *testing.T) {
	// Create a config with HTTP URL
	cfg := &config.Config{
		MQTT: config.MQTTConfig{
			Broker:   "http://mqtt.example.com",
			Port:     1883,
			ClientID: "test-client",
		},
	}

	// Get the broker URL
	brokerURL := cfg.GetMQTTBrokerURL()

	// Verify it's converted to TCP protocol
	if !strings.HasPrefix(brokerURL, "tcp://") {
		t.Errorf("Expected broker URL to start with tcp://, got %s", brokerURL)
	}
}
