package device

import (
	"timechain-gateway/internal/config"
	"timechain-gateway/pkg/models"
)

type Manager struct {
	devices []models.Device
}

func NewManager(cfg *config.Config) *Manager {
	// 初始化设备列表
	return &Manager{
		devices: []models.Device{
			// 从配置加载设备
			{ID: "device1", Type: "temperature"},
			{ID: "device2", Type: "humidity"},
		},
	}
}

func (m *Manager) GetDevices() []models.Device {
	return m.devices
}
