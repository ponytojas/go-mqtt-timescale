package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ponytojas/go-mqtt-timescale/config"
	"github.com/ponytojas/go-mqtt-timescale/internal/database"
	"github.com/ponytojas/go-mqtt-timescale/internal/mqtt"
)

func main() {
	log.Println("Starting MQTT to TimescaleDB service...")

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Printf("Error loading config: %v. Using default configuration.", err)
		cfg = config.GetDefaultConfig()
	}

	// Initialize database connection
	log.Println("Connecting to TimescaleDB...")
	db, err := database.NewTimescaleDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize table
	log.Println("Initializing database table...")
	if err := db.InitializeTable(); err != nil {
		log.Fatalf("Failed to initialize table: %v", err)
	}

	// Initialize MQTT client
	log.Println("Setting up MQTT client...")
	mqttClient, err := mqtt.NewClient(cfg, db)
	if err != nil {
		log.Fatalf("Failed to create MQTT client: %v", err)
	}

	// Connect to MQTT broker
	if err := mqttClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()

	// Subscribe to topic
	if err := mqttClient.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	log.Printf("Service is running. Subscribed to topic: %s", cfg.MQTT.Topic)

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down...")
}
