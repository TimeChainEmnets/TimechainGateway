package main

import (
  "log"
  "timechain-gateway/internal/blockchain"
  "timechain-gateway/internal/config"
  "timechain-gateway/internal/data"
  "timechain-gateway/internal/device"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	deviceManager := device.NewManager(cfg)
	dataCollector := data.NewCollector(deviceManager)
	dataProcessor := data.NewProcessor(cfg)
	blockchainClient := blockchain.NewClient(cfg)

	gateway := NewGateway(cfg, dataCollector, dataProcessor, blockchainClient)
	gateway.Start()
}

type Gateway struct {
	config           *config.Config
	dataCollector    *data.Collector
	dataProcessor    *data.Processor
	blockchainClient *blockchain.Client
}

func NewGateway(cfg *config.Config, collector *data.Collector, processor *data.Processor, client *blockchain.Client) *Gateway {
	return &Gateway{
		config:           cfg,
		dataCollector:    collector,
		dataProcessor:    processor,
		blockchainClient: client,
	}
}

func (g *Gateway) Start() {
	for {
		rawData := g.dataCollector.Collect()
		processedData := g.dataProcessor.Process(rawData)
		g.blockchainClient.SendData(processedData)
	}
}
