package models

import (
	"time"
)

type SensorData struct {
	Timestamp   time.Time `json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Light       float64   `json:"light"`
	Device_ID    string    `json:"device_id"`
}
