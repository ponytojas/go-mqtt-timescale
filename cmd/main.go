package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ponytojas/go-mqtt-timescale/config"
	"github.com/ponytojas/go-mqtt-timescale/internal/mqtt"
	"github.com/ponytojas/go-mqtt-timescale/internal/supabase"
)

func main() {
	log.Println("Starting MQTT to Supabase service...")

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Printf("Error loading config: %v. Using default configuration.", err)
		cfg = config.GetDefaultConfig()
	}

	log.Println("Configuring Supabase Data API client...")
	supabaseClient, err := supabase.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to configure Supabase client: %v", err)
	}

	// Initialize MQTT client
	log.Println("Setting up MQTT client...")
	mqttClient, err := mqtt.NewClient(cfg, supabaseClient)
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
