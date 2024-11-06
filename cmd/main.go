package main

import (
	"bytes"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"timechain-gateway/internal/blockchain"
	"timechain-gateway/internal/config"
	"timechain-gateway/internal/data"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建 MQTT 服务器
	server := mqtt.New(nil)

	// 添加认证钩子（如果需要）
	if err = server.AddHook(new(auth.AllowHook), nil); err != nil {
		log.Fatalf("Failed to add auth hook: %v", err)
	}
	// Add connection hook before starting server
	if err := server.AddHook(new(ConnectionEstablishedHook), nil); err != nil {
		log.Fatalf("Failed to add connection hook: %v", err)
	}
	// 设置 TCP 监听器
	tcpConfig := listeners.Config{
		ID:      "t1",
		Address: cfg.MQTT.Address,
	}

	tcp := listeners.NewTCP(tcpConfig)
	if err = server.AddListener(tcp); err != nil {
		log.Fatalf("Failed to add TCP listener: %v", err)
	}

	// 创建一个 WaitGroup 来等待所有 goroutine 完成
	var wg sync.WaitGroup
	// 启动 MQTT 服务器
	wg.Add(1)

	// 启动服务器
	go func() {
		defer wg.Done()
		if err := server.Serve(); err != nil {
			log.Fatalf("Failed to start MQTT server: %v", err)
		}
	}()

	dataCollector := data.NewCollector()
	dataProcessor := data.NewProcessor(cfg)
	blockchainClient := blockchain.NewClient(cfg)

	gateway := NewGateway(cfg, dataCollector, dataProcessor, blockchainClient)

	// 添加消息处理钩子
	messageHook := &MessageHook{
		collector: dataCollector,
	}
	if err := server.AddHook(messageHook, nil); err != nil {
		log.Fatalf("Failed to add message hook: %v", err)
	}

	// 启动 Gateway
	wg.Add(1)
	go func() {
		defer wg.Done()
		gateway.Start()
	}()

	// 等待中断信号以优雅地关闭服务器
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	// 收到信号后，开始优雅关闭
	log.Println("Shutting down...")

	// 收到中断信号后关闭服务器
	server.Close()

	gateway.Stop() // 假设您有一个 Stop 方法来清理 Gateway
	// 等待所有 goroutine 完成
	wg.Wait()
	log.Println("Shutdown complete")
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
	// 启动collector
	g.dataCollector.StartCollecting()

	// 启动处理循环
	go func() {
		processChan := g.dataCollector.GetProcessedChannel()
		batchNum := g.dataCollector.BigBatchSize / g.dataCollector.BatchSize
		for bigBatch := range processChan {
			// 处理数据
			processedData := g.dataProcessor.ProcessBatch(bigBatch)
			// 访问区块链RPC节点调用合约，交易上链
			// 将数据发送到存储节点
			if err := g.blockchainClient.SendData(processedData, batchNum); err != nil {
				log.Printf("Error sending data to blockchain: %v", err)
			}
		}
	}()
}

func (g *Gateway) Stop() {
	g.dataCollector.Stop()
}

// ConnectionEstablishedHook 是一个钩子，当客户端连接到服务器时调用
type ConnectionEstablishedHook struct {
	mqtt.HookBase
}

// Init initializes the hook
func (h *ConnectionEstablishedHook) Init(config interface{}) error {
	return nil
}

// Provides indicates which hook methods this hook provides
func (h *ConnectionEstablishedHook) Provides(b byte) bool {
	return b == mqtt.OnConnect
}

// OnConnect is called when a client connects to the server
func (h *ConnectionEstablishedHook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	log.Printf("Client connected: %s", cl.ID)
	return nil
}

type MessageHook struct {
	mqtt.HookBase
	collector *data.Collector
}

func (h *MessageHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnect,
		mqtt.OnPublish,
	}, []byte{b})
}

func (h *MessageHook) OnConnect(cl *mqtt.Client, pk packets.Packet) error {
	log.Printf("Client connected: %s", cl.ID)
	return nil
}

func (h *MessageHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	return pk, h.collector.HandleMQTTMessage(pk.TopicName, pk.Payload)
}
