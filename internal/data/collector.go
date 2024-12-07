package data

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"timechain-gateway/internal/config"
	"timechain-gateway/pkg/models"
)

type Collector struct {
	config          *config.Config							// 设备配置
	dataChan        chan models.SensorData	// device => collector 接收MQTT数据
	processedChan   chan []models.SensorData // collector => processor 发送处理后的数据
	BatchSize       int	// 通过mqtt发送的单个数据包的数据量
	BigBatchSize    int	// 每次对进行graph更新的数据量
	processInterval time.Duration	// 更新Graph的时间间隔
	// mutex           sync.RWMutex	// 用于保护graph的读写操作
	stopChan        chan struct{}
}

func NewCollector(cfg *config.Config) *Collector {
	return &Collector{
		config:          cfg,
		dataChan:        make(chan models.SensorData, cfg.DeviceConfig.ProcessInterval/cfg.DeviceConfig.ScanInterval),
		processedChan:   make(chan []models.SensorData, 10),
		BatchSize:       8,
		BigBatchSize:    cfg.DeviceConfig.DeviceNumber * cfg.DeviceConfig.ProcessInterval / cfg.DeviceConfig.ScanInterval,
		processInterval: time.Second * time.Duration(cfg.DeviceConfig.ProcessInterval),
		stopChan:        make(chan struct{}),
	}
}

// HandleMQTTMessage 处理来自 MQTT 的消息
func (c *Collector) HandleMQTTMessage(topic string, payload []byte) error {
	// 解析消息
	var data models.SensorData
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	// 非阻塞方式发送数据
	select {
	case c.dataChan <- data:
		return nil
	default:
		return fmt.Errorf("channel buffer full, message dropped")
	}
}

// StartCollecting 开始通过mqtt协议接收device传来的数据
func (c *Collector) StartCollecting() {
	go func() {
		// 收集满 batch size个数据后，发送给processBatch处理
		BigBatch := make([]models.SensorData, 0, c.BigBatchSize)
		ticker := time.NewTicker(c.processInterval)
		// 若未填满数据，触发ticker后直接发送给processor模块处理
		defer ticker.Stop()
		for {
			select {
			case data := <-c.dataChan:
				BigBatch = append(BigBatch, data)
				if len(BigBatch) >= c.BigBatchSize {
					c.sendBatchToProcessor(BigBatch)
					BigBatch = make([]models.SensorData, 0, c.BigBatchSize)
				}
			case <-ticker.C:
				if len(BigBatch) > 0 {
					c.sendBatchToProcessor(BigBatch)
					BigBatch = make([]models.SensorData, 0, c.BigBatchSize)
				}
			case <-c.stopChan:
				if len(BigBatch) > 0 {
					c.sendBatchToProcessor(BigBatch)
					BigBatch = make([]models.SensorData, 0, c.BigBatchSize)
				}
				return
			}
		}
	}()
	log.Println("Collector started with batch processing")
}

func (c *Collector) sendBatchToProcessor(BigBatch []models.SensorData) {
	c.processedChan <- BigBatch
}

// GetProcessedChannel 返回用于处理数据的channel
func (c *Collector) GetProcessedChannel() <-chan []models.SensorData {
	return c.processedChan
}

// Stop 优雅停止收集器
func (c *Collector) Stop() {
	close(c.stopChan)
}
