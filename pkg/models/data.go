package models

import (
	"math/rand"
	"time"
)

type Device struct {
	ID   string
	Type string
}

type RawData struct {
	DeviceID  string
	Timestamp int64
	Value     float64
}

type ProcessedData struct {
	Batches []DataBatch
}

type DataBatch struct {
	Timestamp int64
	DataHash  []byte
	Clusters  []ClusterInfo
}

type ClusterInfo struct {
	Center []float64
	Size   int
}

// 添加 CollectData 方法到 Device 结构体
func (d *Device) CollectData() RawData {
	// 这里应该实现实际的数据收集逻辑
	// 现在我们只是生成一些模拟数据
	return RawData{
		DeviceID:  d.ID,
		Timestamp: time.Now().Unix(),
		Value:     rand.Float64() * 100, // 生成0-100之间的随机数
	}
}
