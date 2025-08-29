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

	viper.SetDefault("database.host", defaultConfig.Database.Host)
	viper.SetDefault("database.port", defaultConfig.Database.Port)
	viper.SetDefault("database.user", defaultConfig.Database.User)
	viper.SetDefault("database.password", defaultConfig.Database.Password)
	viper.SetDefault("database.dbname", defaultConfig.Database.DBName)
	viper.SetDefault("database.sslmode", defaultConfig.Database.SSLMode)

	viper.SetDefault("timescale.table_name", defaultConfig.Timescale.TableName)

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

	// Database configuration
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "DATABASE_SSLMODE")

	// Timescale configuration
	viper.BindEnv("timescale.table_name", "TIMESCALE_TABLE_NAME")

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
	// log the URI
	log.Printf("Connecting to database at 'host=%s port=%d user=%s dbname=%s sslmode=%s'",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.DBName,
		c.Database.SSLMode,
	)
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
