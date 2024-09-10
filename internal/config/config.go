package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DeviceConfig     DeviceConfig     `json:"device_config"`
	BlockchainConfig BlockchainConfig `json:"blockchain_config"`
	StorageConfig    StorageConfig    `json:"storage_config"`
}

type DeviceConfig struct {
	ScanInterval int `json:"scan_interval"` // 设备扫描间隔（秒）
}

type BlockchainConfig struct {
	NodeURL    string `json:"node_url"`
	ChainID    string `json:"chain_id"`
	GasLimit   uint64 `json:"gas_limit"`
	GasPrice   string `json:"gas_price"`
	WalletSeed string `json:"wallet_seed"`
}

type StorageConfig struct {
	DataDir string `json:"data_dir"` // 本地数据存储目录
}

func Load() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
