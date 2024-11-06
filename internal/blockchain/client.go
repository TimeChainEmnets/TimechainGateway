package blockchain

import (
	"fmt"
	"timechain-gateway/internal/config"
	"timechain-gateway/pkg/models"
)

type Client struct {
	config *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{config: cfg}
}

func (c *Client) SendData(data [][]models.SensorData, batchNum int) error {
	// 实现与 TimeChain 区块链的交互逻辑
	// 发送处理后的数据到区块链
	fmt.Printf("Sending data to TimeChain blockchain... %d", batchNum)

	return nil
}
