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
	dataChan  chan models.RawData
	batchSize int
	mutex     sync.RWMutex
	stopChan  chan struct{}
}

func NewCollector() *Collector {
	return &Collector{
		dataChan:  make(chan models.RawData, 100),
		batchSize: 10,
		stopChan:  make(chan struct{}),
	}
}

// HandleMQTTMessage 处理来自 MQTT 的消息
func (c *Collector) HandleMQTTMessage(topic string, payload []byte) error {
	// 解析消息
	var data models.RawData
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
		batch := make([]models.RawData, 0, c.batchSize)
		ticker := time.NewTicker(time.Second) // 定时刷新
		defer ticker.Stop()
		for {
			select {
			case data := <-c.dataChan:
				batch = append(batch, data)
				if len(batch) >= c.batchSize {
					c.processBatch(batch)
					batch = batch[:0] // 清空批次
				}
			case <-ticker.C:
				if len(batch) > 0 {
					c.processBatch(batch)
					batch = batch[:0]
				}
			case <-c.stopChan:
				if len(batch) > 0 {
					c.processBatch(batch)
				}
				return
			}
		}
	}()
	log.Println("Collector started with batch processing")
}

func (c *Collector) processBatch(batch []models.RawData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	log.Printf("Processing batch of %d messages\n", len(batch))
	for _, data := range batch {
		log.Printf("Processing data: %+v\n", data)
	}
}

// Stop 优雅停止收集器
func (c *Collector) Stop() {
	close(c.stopChan)
}
