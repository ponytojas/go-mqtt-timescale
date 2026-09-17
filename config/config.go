package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	MQTT     MQTTConfig     `mapstructure:"mqtt"`
	Supabase SupabaseConfig `mapstructure:"supabase"`
}

// MQTTConfig holds MQTT connection configuration
type MQTTConfig struct {
	Broker   string `mapstructure:"broker"`
	Port     int    `mapstructure:"port"`
	ClientID string `mapstructure:"client_id"`
	Topic    string `mapstructure:"topic"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// SupabaseConfig holds the Data REST API settings and target table.
type SupabaseConfig struct {
	URL       string `mapstructure:"url"`
	APIKey    string `mapstructure:"api_key"`
	Schema    string `mapstructure:"schema"`
	TableName string `mapstructure:"table_name"`
}

// LoadConfig loads configuration from file and/or environment variables
func LoadConfig(path string) (*Config, error) {
	// Set default values first (lowest precedence)
	defaultConfig := GetDefaultConfig()
	viper.SetDefault("mqtt.broker", defaultConfig.MQTT.Broker)
	viper.SetDefault("mqtt.port", defaultConfig.MQTT.Port)
	viper.SetDefault("mqtt.client_id", defaultConfig.MQTT.ClientID)
	viper.SetDefault("mqtt.topic", defaultConfig.MQTT.Topic)
	viper.SetDefault("mqtt.username", defaultConfig.MQTT.Username)
	viper.SetDefault("mqtt.password", defaultConfig.MQTT.Password)

	viper.SetDefault("supabase.url", defaultConfig.Supabase.URL)
	viper.SetDefault("supabase.api_key", defaultConfig.Supabase.APIKey)
	viper.SetDefault("supabase.schema", defaultConfig.Supabase.Schema)
	viper.SetDefault("supabase.table_name", defaultConfig.Supabase.TableName)

	// Try to load from config file (medium precedence)
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set up environment variable support (highest precedence)
	viper.SetEnvPrefix("") // No prefix
	// Keep backward compatibility with MQTT_BROKER_URL
	viper.BindEnv("mqtt.broker", "MQTT_BROKER_URL")

	// Map all configuration keys to environment variables
	// Example: mqtt.broker -> MQTT_BROKER
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Explicitly bind all environment variables to ensure they work
	// MQTT configuration
	viper.BindEnv("mqtt.broker", "MQTT_BROKER")
	viper.BindEnv("mqtt.port", "MQTT_PORT")
	viper.BindEnv("mqtt.client_id", "MQTT_CLIENT_ID")
	viper.BindEnv("mqtt.topic", "MQTT_TOPIC")
	viper.BindEnv("mqtt.username", "MQTT_USERNAME")
	viper.BindEnv("mqtt.password", "MQTT_PASSWORD")

	// Supabase configuration
	viper.BindEnv("supabase.url", "SUPABASE_URL")
	viper.BindEnv("supabase.api_key", "SUPABASE_API_KEY")
	viper.BindEnv("supabase.schema", "SUPABASE_SCHEMA")
	viper.BindEnv("supabase.table_name", "SUPABASE_TABLE_NAME")

	// Try to read config file, but don't fail if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			log.Printf("Warning: error reading config file: %v", err)
		} else {
			log.Println("No config file found, using environment variables and defaults")
		}
		// We'll continue with environment variables and defaults
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
	return &Config{
		MQTT: MQTTConfig{
			Broker:   "https://mqtt.ponytojas.dev", // Updated default
			Port:     8883,                         // Updated default port for TLS
			ClientID: "go-mqtt-client",
			Topic:    "sensor/#",
			Username: "",
			Password: "",
		},
		Supabase: SupabaseConfig{
			Schema:    "labtools",
			TableName: "sensor_data",
		},
	}
}

// GetMQTTBrokerURL returns the MQTT broker URL
func (c *Config) GetMQTTBrokerURL() string {
	brokerURL := c.MQTT.Broker

	// If the URL already has a protocol, use it as is
	if strings.HasPrefix(brokerURL, "tcp://") ||
		strings.HasPrefix(brokerURL, "ssl://") ||
		strings.HasPrefix(brokerURL, "ws://") ||
		strings.HasPrefix(brokerURL, "wss://") {
		// If there's no port in the URL, add the default port
		if !strings.Contains(brokerURL[6:], ":") {
			brokerURL = fmt.Sprintf("%s:%d", brokerURL, c.MQTT.Port)
		}
		return brokerURL
	}

	// Handle http:// and https:// protocols by converting to mqtt protocols
	if strings.HasPrefix(brokerURL, "http://") {
		host := brokerURL[7:]
		if !strings.Contains(host, ":") {
			host = fmt.Sprintf("%s:%d", host, c.MQTT.Port)
		}
		return fmt.Sprintf("tcp://%s", host)
	}

	if strings.HasPrefix(brokerURL, "https://") {
		host := brokerURL[8:]
		if !strings.Contains(host, ":") {
			host = fmt.Sprintf("%s:%d", host, c.MQTT.Port)
		}
		return fmt.Sprintf("ssl://%s", host)
	}

	// If no protocol is specified, use tcp:// with the configured port
	log.Printf("No protocol specified in broker URL '%s', defaulting to tcp://", brokerURL)
	return fmt.Sprintf("tcp://%s:%d", brokerURL, c.MQTT.Port)
}
