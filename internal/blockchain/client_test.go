package blockchain

import (
	"testing"
	"timechain-gateway/internal/config"
	"timechain-gateway/pkg/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestGetBalance(t *testing.T) {
	// 初始化测试配置和客户端
	cfg := &config.Config{
		// 设置配置项
	}
	client := NewClient(cfg)

	// 测试 GetBalance 方法
	balance, err := client.GetBalance()
	assert.NoError(t, err)
	assert.NotNil(t, balance)

	// 可以添加更多断言来验证 balance 的值
}

func TestClose(t *testing.T) {
	// 初始化测试配置和客户端
	cfg, err := config.Load("../../config.json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	client := NewClient(cfg)

	// 测试 Close 方法
	client.Close()

	// 可以添加断言来验证 ethclient 是否已关闭
}

func TestGetLSH(t *testing.T) {
	// 测试数据
	data := [][]models.SensorData{
		{
			{Timestamp: 1, Value: 36.3, DeviceID: "device001", SensorID: "sensor001", Type: "temperature", Unit: "Celsius"},
			{Timestamp: 2, Value: 36.1, DeviceID: "device001", SensorID: "sensor001", Type: "temperature", Unit: "Celsius"},
		},
	}

	// 预期结果
	expectedRootHashes := []common.Hash{common.Hash{}}

	// 测试 getLSH 方法
	rootHashes := getLSH(data)
	assert.Equal(t, expectedRootHashes, rootHashes)

	// 可以添加更多测试用例来验证 getLSH 方法的正确性
}
