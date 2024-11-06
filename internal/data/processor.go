package data

import (
	"log"
	"sync"
	"timechain-gateway/internal/config"
	"timechain-gateway/pkg/models"
)

type Processor struct {
	config *config.Config
	mutex  sync.RWMutex
}



func NewProcessor(cfg *config.Config) *Processor {
	return &Processor{config: cfg}
}

func (p *Processor) ProcessBatch(batch []models.SensorData) []models.SensorData {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Printf("Processing batch of %d messages\n", len(batch))
	for _, data := range batch {
		log.Printf("Processing data: %+v\n", data)
		// TODO:build a Graph
		// TODO:将batchsize个
	}

	return batch
}
type SensorData struct {
	SensorID string `json:"sensor_id"`
	DeviceID  string  `json:"device_id"`
	Timestamp int64   `json:"timestamp"`
	Type      string  `json:"type"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
}