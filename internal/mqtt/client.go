package mqtt

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ponytojas/go-mqtt-timescale/config"
	"github.com/ponytojas/go-mqtt-timescale/internal/database"
	"github.com/ponytojas/go-mqtt-timescale/internal/models"
)

// Client handles MQTT connection and message processing
type Client struct {
	client   mqtt.Client
	db       *database.TimescaleDB
	config   *config.Config
	stopChan chan struct{}
}

// NewClient creates a new MQTT client
func NewClient(cfg *config.Config, db *database.TimescaleDB) (*Client, error) {
	opts := mqtt.NewClientOptions()
	brokerURL := cfg.GetMQTTBrokerURL()
	opts.AddBroker(brokerURL)
	opts.SetClientID(cfg.MQTT.ClientID)

	// Configure TLS if using SSL or HTTPS
	if strings.HasPrefix(brokerURL, "ssl://") || strings.HasPrefix(brokerURL, "wss://") {
		log.Printf("Configuring TLS for secure connection to %s", brokerURL)
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		opts.SetTLSConfig(tlsConfig)
	}

	if cfg.MQTT.Username != "" {
		opts.SetUsername(cfg.MQTT.Username)
		opts.SetPassword(cfg.MQTT.Password)
	}

	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})
	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		log.Println("Attempting to reconnect to MQTT broker...")
	})

	client := mqtt.NewClient(opts)
	return &Client{
		client:   client,
		db:       db,
		config:   cfg,
		stopChan: make(chan struct{}),
	}, nil
}

// Connect connects to the MQTT broker
func (c *Client) Connect() error {
	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	log.Printf("Connected to MQTT broker: %s", c.config.GetMQTTBrokerURL())
	return nil
}

// Subscribe subscribes to the configured topic
func (c *Client) Subscribe() error {
	handler := func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))
		c.processMessage(msg.Payload())
	}

	token := c.client.Subscribe(c.config.MQTT.Topic, 0, handler)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", c.config.MQTT.Topic, token.Error())
	}
	log.Printf("Subscribed to topic: %s", c.config.MQTT.Topic)
	return nil
}

// Disconnect disconnects from the MQTT broker
func (c *Client) Disconnect() {
	c.client.Disconnect(250)
	log.Println("Disconnected from MQTT broker")
}

// Stop stops the client
func (c *Client) Stop() {
	close(c.stopChan)
}

// WaitForStop waits for the client to stop
func (c *Client) WaitForStop() {
	<-c.stopChan
}

// processMessage processes an MQTT message and stores it in the database
func (c *Client) processMessage(payload []byte) {
	var rawData map[string]interface{}
	if err := json.Unmarshal(payload, &rawData); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	// Parse timestamp
	var timestamp time.Time
	if tsStr, ok := rawData["timestamp"].(string); ok {
		var err error
		timestamp, err = time.Parse(time.RFC3339, tsStr)
		if err != nil {
			log.Printf("Error parsing timestamp: %v", err)
			timestamp = time.Now() // Fallback to current time
		}
	} else {
		timestamp = time.Now() // Fallback to current time
	}

	// Extract sensor values
	temperature, _ := getFloat64Value(rawData, "temperature")
	humidity, _ := getFloat64Value(rawData, "humidity")
	light, _ := getFloat64Value(rawData, "light")
	device_id, ok := rawData["device_id"].(string)
	if !ok {
		log.Println("Error: device_id is missing or not a string")
		return
	}

	// Create sensor data
	sensorData := &models.SensorData{
		Timestamp:   timestamp,
		Temperature: temperature,
		Humidity:    humidity,
		Light:       light,
		Device_ID:   device_id,
	}

	// Insert into database
	if err := c.db.InsertSensorData(sensorData); err != nil {
		log.Printf("Error inserting sensor data: %v", err)
		return
	}

	log.Printf("Successfully processed and stored sensor data: time=%v, temp=%.2f, humidity=%.2f, light=%.2f",
		timestamp, temperature, humidity, light)
}

// getFloat64Value safely extracts a float64 value from the map
func getFloat64Value(data map[string]interface{}, key string) (float64, bool) {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v, true
		case string:
			if f, err := parseFloat(v); err == nil {
				return f, true
			}
		case int:
			return float64(v), true
		case int64:
			return float64(v), true
		}
	}
	return 0, false
}

// parseFloat attempts to parse a string as a float64
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
