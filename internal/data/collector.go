package data

import (
	"timechain-gateway/internal/device"
	"timechain-gateway/pkg/models"
)

type Collector struct {
	deviceManager *device.Manager
}

func NewCollector(dm *device.Manager) *Collector {
	return &Collector{deviceManager: dm}
}

func (c *Collector) Collect() []models.RawData {
	var rawData []models.RawData
	for _, device := range c.deviceManager.GetDevices() {
		// 从每个设备收集数据
		data := device.CollectData()
		rawData = append(rawData, data)
	}
	return rawData
}
