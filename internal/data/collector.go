package data

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"timechain-gateway/pkg/models"
)

type Collector struct {
	dataChan      chan models.SensorData
	processedChan chan []models.SensorData // 新增：用于发送给processor的channel
	batchSize     int
	interval      time.Duration
	mutex         sync.RWMutex
	stopChan      chan struct{}
}

func NewCollector() *Collector {
	return &Collector{
		dataChan:      make(chan models.SensorData, 100),
		processedChan: make(chan []models.SensorData, 10),
		batchSize:     100,
		interval:      time.Second * 30,
		stopChan:      make(chan struct{}),
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

// StartCollecting 开始收集数据
func (c *Collector) StartCollecting() {
	go func() {
		// 收集满 batch size个数据后，发送给processBatch处理
		// batch size暂定100，10个设备，每个设备10个数据
		batch := make([]models.SensorData, 0, c.batchSize)
		ticker := time.NewTicker(c.interval) // 定时刷新
		defer ticker.Stop()
		for {
			select {
			case data := <-c.dataChan:
				batch = append(batch, data)
				if len(batch) >= c.batchSize {
					c.sendBatchToProcessor(batch)
					batch = make([]models.SensorData, 0, c.batchSize)
				}
			case <-ticker.C:
				if len(batch) > 0 {
					c.sendBatchToProcessor(batch)
					batch = make([]models.SensorData, 0, c.batchSize)
				}
			case <-c.stopChan:
				if len(batch) > 0 {
					c.sendBatchToProcessor(batch)
					batch = make([]models.SensorData, 0, c.batchSize)
				}
				return
			}
		}
	}()
	log.Println("Collector started with batch processing")
}

func (c *Collector) sendBatchToProcessor(batch []models.SensorData) {
	c.processedChan <- batch
}

// GetProcessedChannel 返回用于处理数据的channel
func (c *Collector) GetProcessedChannel() <-chan []models.SensorData {
	return c.processedChan
}

// Stop 优雅停止收集器
func (c *Collector) Stop() {
	close(c.stopChan)
}
