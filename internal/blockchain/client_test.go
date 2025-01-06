package blockchain

import (
	"fmt"
	"testing"

	"github.com/TimeChainEmnets/TimechainGateway/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetBalance(t *testing.T) {
	// 初始化测试配置和客户端
	cfg := &config.Config{
		// 设置配置项
		BlockchainConfig: config.BlockchainConfig{
			NodeURL:         "https://sepolia.infura.io/v3/b5e6f67f813a41f983b6c9f3f5bba595", // 本地测试节点
			ChainID:         11155111,                                                        // 测试网络ID
			GasLimit:        3000000,
			ContractAddress: "0xA5dd2cf9d382c83fB0e6Fb218fF6777CfD132DF2",
			PrivateKey:      "9b55a066922e55fba996585f825498f246180d4392eb4c79c3a0b7d82f762fff",
		},
	}
	client := NewClient(cfg)

	// 测试 GetBalance 方法
	balance, err := client.GetBalance()
	fmt.Println(balance)
	assert.NoError(t, err)
	assert.NotNil(t, balance)

	// 可以添加更多断言来验证 balance 的值
}
