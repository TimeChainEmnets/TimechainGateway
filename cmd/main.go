package main

import (
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// 创建 MQTT 服务器
	server := mqtt.New(nil)

	// 添加认证钩子（如果需要）
	err = server.AddHook(new(auth.AllowHook), nil)
	if err != nil {
		log.Fatalf("Failed to add auth hook: %v", err)
	}
	// 设置 TCP 监听器
	tcpConfig := listeners.Config{
		ID:      "t1",
		Address: cfg.MQTT.Address,
	}
	tcp := listeners.NewTCP(tcpConfig)
	err = server.AddListener(tcp)
	if err != nil {
		log.Fatalf("Failed to add TCP listener: %v", err)
	}

	// 启动服务器
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatalf("Failed to start MQTT server: %v", err)
		}
	}()

	deviceManager := device.NewManager(cfg, server)
	dataCollector := data.NewCollector(deviceManager)
	dataProcessor := data.NewProcessor(cfg)
	blockchainClient := blockchain.NewClient(cfg)

	gateway := NewGateway(cfg, dataCollector, dataProcessor, blockchainClient)
	// 启动 Gateway
	go gateway.Start()

	// 等待中断信号以优雅地关闭服务器
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	// 收到中断信号后关闭服务器
	server.Close()
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
