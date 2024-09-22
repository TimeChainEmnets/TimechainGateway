package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Server           ServerConfig     `json:"server_config"`
	MQTT             MQTTConfig       `json:"mqtt_config"`
	DeviceConfig     DeviceConfig     `json:"device_config"`
	BlockchainConfig BlockchainConfig `json:"blockchain_config"`
	StorageConfig    StorageConfig    `json:"storage_config"`
}

type ServerConfig struct {
	Port int `json:"port"`
}

type MQTTConfig struct {
	Address         string `json:"address"`
	DeviceInfoTopic string `json:"device_info_topic"`
}

type DeviceConfig struct {
	ScanInterval int `json:"scan_interval"` // 设备扫描间隔（秒）
}

type BlockchainConfig struct {
	NodeURL         string `json:"node_url"`
	ChainID         string `json:"chain_id"`
	GasLimit        uint64 `json:"gas_limit"`
	ContractAddress string `json:"contract_address"`
}

type StorageConfig struct {
	DataDir string `json:"storage_dir"` // 本地数据存储目录
}

func Load(fileName string) (*Config, error) {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// 构造配置文件的相对路径
	configPath := filepath.Join(currentDir, fileName)

	file, err := os.Open(configPath)
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
