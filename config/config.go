package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	MQTT      MQTTConfig      `mapstructure:"mqtt"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Timescale TimescaleConfig `mapstructure:"timescale"`
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

// DatabaseConfig holds Postgres connection configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// TimescaleConfig holds Timescale specific configuration
type TimescaleConfig struct {
	TableName string `mapstructure:"table_name"`
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set up environment variable mappings
	viper.SetEnvPrefix("") // No prefix
	viper.BindEnv("mqtt.broker", "MQTT_BROKER_URL")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
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
			Topic:    "sensors/data",
			Username: "",
			Password: "",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "iot_data",
			SSLMode:  "disable",
		},
		Timescale: TimescaleConfig{
			TableName: "sensor_data",
		},
	}
}

// GetDBConnString returns the database connection string
func (c *Config) GetDBConnString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
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