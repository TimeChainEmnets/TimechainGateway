package models

import (
	"math/rand"
	"time"
)

type Device struct {
	ID   string
	Type string
}

type SensorData struct {
	SensorID string `json:"sensor_id"`
	DeviceID  string  `json:"device_id"`
	Timestamp int64   `json:"timestamp"`
	Type      string  `json:"type"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
}
