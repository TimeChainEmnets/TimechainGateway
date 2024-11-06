package data

import (
	"fmt"
	"log"
	"sync"
	"time"
	"timechain-gateway/internal/config"
	"timechain-gateway/pkg/models"

	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/mat"
)

type Processor struct {
	config       *config.Config
	mutex        sync.RWMutex
	graph        *simple.WeightedUndirectedGraph
	nodeMap      map[string]int64 // deviceID -> node ID mapping
	bigBatchSize int
}

func NewProcessor(cfg *config.Config) *Processor {
	return &Processor{
		config:       cfg,
		graph:        simple.NewWeightedUndirectedGraph(0, 0),
		nodeMap:      make(map[string]int64),
		BigBatchSize: cfg.DeviceConfig.DeviceNumber * cfg.DeviceConfig.ProcessInterval / cfg.DeviceConfig.ScanInterval,
	}
}

// ProcessBatch 处理数据：将收到的bigBatch构建成多个标准大小的Batch
// 根据bigBatch的长度除以batchSize(20)，得到batch的数量batchNum
func (p *Processor) ProcessBatch(bigBatch []models.SensorData) [][]models.SensorData {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 制造fake data补足数量，以保证bigBatch的长度为batchSize的整数倍
	if len(bigBatch) < p.bigBatchSize {
		needed := p.bigBatchSize - len(bigBatch)
		for i := 0; i < needed; i++ {
			fakeData := models.SensorData{
				SensorID:  bigBatch[0].SensorID,
				DeviceID:  bigBatch[0].DeviceID,
				Timestamp: time.Now().Unix(),
				Type:      bigBatch[0].Type,
				Value:     0.0,
				Unit:      bigBatch[0].Unit,
			}
			bigBatch = append(bigBatch, fakeData)
		}
	}
	batchNum := p.bigBatchSize / 20
	batches := make([][]models.SensorData, batchNum)
	for i := range batches {
		batches[i] = make([]models.SensorData, 20)
	}
	/*
		处理设备数据并更新图结构，batch中的数据可能来源于多个DeviceID
		config中设置设备数量，如需修改设备数量，修改config文件并重启即可
		若当前graph中设备数量小于config中设置的设备数量，则添加新设备，添加完后再处理数据
	*/
	if len(p.nodeMap) < p.config.DeviceConfig.DeviceNumber {
		for _, data := range bigBatch {
			p.addDeviceToGraph(data.DeviceID)
		}
	}
	if p.config.DeviceConfig.DeviceNumber < 4 {
		// 如果设备数量小于4，则无需聚类，直接将DeviceID相同的放在一起构成Batch
		deviceDataMap := make(map[string][]models.SensorData)
		for _, data := range bigBatch {
			deviceDataMap[data.DeviceID] = append(deviceDataMap[data.DeviceID], data)
		}

		batchIndex := 0
		batchPos := 0
		for _, deviceData := range deviceDataMap {
			for len(deviceData) > 0 {
				spaceLeft := 20 - batchPos
				if len(deviceData) >= spaceLeft {
					batches[batchIndex] = append(batches[batchIndex], deviceData[:spaceLeft]...)
					deviceData = deviceData[spaceLeft:]
					batchIndex++
					batchPos = 0
				} else {
					// 如果当前batch未填满，则用不同DeviceID的数据填充
					batches[batchIndex] = append(batches[batchIndex], deviceData...)
					batchPos += len(deviceData)
					deviceData = nil
				}
			}
		}
	} else {
		// 如果设备数量大于等于4，进行聚类，类别数量设定未 DeviceNumber/3
		// classNumber := p.config.DeviceConfig.DeviceNumber / 3
		// TODO1:构造符合高斯分布的Query
		// TODO2:使用Query结合UpdateEdgeWeight更新图的权重
		// TODO3:使用谱聚类算法对图进行聚类，得到类别结果
		// TODO4:根据分类结果对bigBatch进行划分，得到batches

	}
	return batches
}

func (p *Processor) addDeviceToGraph(deviceID string) {
	if _, exists := p.nodeMap[deviceID]; !exists {
		// 添加新节点, nodeID从1开始
		nodeID := int64(len(p.nodeMap) + 1)
		p.nodeMap[deviceID] = nodeID
		// 创建新节点
		newNode := simple.Node(nodeID)
		p.graph.AddNode(newNode)
		// 与现有节点建立带权重的连接
		const initialWeight = 1.0 // 设置初始权重

		// 遍历已存在的节点，建立边连接
		for existingDeviceID, existingNodeID := range p.nodeMap {
			if existingDeviceID != deviceID {
				existingNode := simple.Node(existingNodeID)
				edge := simple.WeightedEdge{
					F: newNode,
					T: existingNode,
					W: initialWeight,
				}
				p.graph.SetWeightedEdge(edge)
			}
		}
		log.Printf("Added new device %s as node %d", deviceID, nodeID)
	}
}

// 获取两个设备之间的权重
func (p *Processor) GetEdgeWeight(deviceID1, deviceID2 string) (float64, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	nodeID1, exists1 := p.nodeMap[deviceID1]
	nodeID2, exists2 := p.nodeMap[deviceID2]

	if !exists1 || !exists2 {
		return 0, fmt.Errorf("device not found")
	}

	weight, ok := p.graph.Weight(nodeID1, nodeID2)
	if !ok {
		return 0, fmt.Errorf("edge not found between devices")
	}
	return weight, nil
}

func (p *Processor) UpdateEdgeWeight(deviceID1, deviceID2 string, weight float64) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	nodeID1, exists1 := p.nodeMap[deviceID1]
	nodeID2, exists2 := p.nodeMap[deviceID2]

	if !exists1 || !exists2 {
		return fmt.Errorf("device not found")
	}

	edge := simple.WeightedEdge{
		F: simple.Node(nodeID1),
		T: simple.Node(nodeID2),
		W: weight,
	}
	p.graph.SetWeightedEdge(edge)

	return nil
}

// 获取邻接矩阵
func (p *Processor) GetAdjacencyMatrix() *mat.Dense {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	n := len(p.nodeMap)
	matrix := mat.NewDense(n, n, nil)

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if p.graph.HasEdgeBetween(int64(i+1), int64(j+1)) {
				matrix.Set(i, j, 1)
			}
		}
	}

	return matrix
}

// 获取拉普拉斯矩阵 (用于谱聚类)
func (p *Processor) GetLaplacianMatrix() *mat.Dense {
	adj := p.GetAdjacencyMatrix()
	n, _ := adj.Dims()

	degree := mat.NewDense(n, n, nil)
	laplacian := mat.NewDense(n, n, nil)

	// 计算度矩阵
	for i := 0; i < n; i++ {
		sum := 0.0
		for j := 0; j < n; j++ {
			sum += adj.At(i, j)
		}
		degree.Set(i, i, sum)
	}

	// L = D - A
	laplacian.Sub(degree, adj)
	return laplacian
}

type SensorData struct {
	SensorID  string  `json:"sensor_id"`
	DeviceID  string  `json:"device_id"`
	Timestamp int64   `json:"timestamp"`
	Type      string  `json:"type"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
}
