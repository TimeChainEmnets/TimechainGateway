package data

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
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
	BatchSize 	 int // 最终单笔交易上链的数据量
}

type SensorData struct {
	SensorID  string  `json:"sensor_id"`
	DeviceID  string  `json:"device_id"`
	Timestamp int64   `json:"timestamp"`
	Type      string  `json:"type"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
}

func NewProcessor(cfg *config.Config) *Processor {
	return &Processor{
		config:       cfg,
		graph:        simple.NewWeightedUndirectedGraph(0, 0),
		nodeMap:      make(map[string]int64),
		bigBatchSize: cfg.DeviceConfig.DeviceNumber * cfg.DeviceConfig.ProcessInterval / cfg.DeviceConfig.ScanInterval,
		BatchSize: 	32,
	}
}

// ProcessBatch 处理数据：将收到的bigBatch构建成多个标准大小的Batch
// 根据bigBatch的长度除以batchSize(32)，得到batch的数量batchNum
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
	batchNum := p.bigBatchSize / p.BatchSize
	batches := make([][]models.SensorData, batchNum)
	for i := range batches {
		batches[i] = make([]models.SensorData, p.BatchSize)
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
	if p.config.DeviceConfig.DeviceNumber < 10 {
		// 如果设备数量小于10，则无需聚类，直接将DeviceID相同的放在一起构成Batch
		deviceDataMap := make(map[string][]models.SensorData)
		for _, data := range bigBatch {
			deviceDataMap[data.DeviceID] = append(deviceDataMap[data.DeviceID], data)
		}

		batchIndex := 0
		batchPos := 0
		for _, deviceData := range deviceDataMap {
			for len(deviceData) > 0 {
				spaceLeft := p.BatchSize - batchPos
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
		// 如果设备数量大于等于10，进行聚类，类别数量设定未 DeviceNumber/3
		classNumber := p.config.DeviceConfig.DeviceNumber / 3
		// TODO1:构造符合高斯分布的Query
		// query共2个参数(deviceID, time)
		deviceIDs := make([]string, len(bigBatch))
		timestamps := make([]int64, len(bigBatch))
		var minDeviceID, maxDeviceID string
		var minTimestamp, maxTimestamp int64
		minTimestamp = 1<<63 - 1 // int64的最大值
		maxTimestamp = 0         // timestamp > 0
		minDeviceID = bigBatch[0].DeviceID
		maxDeviceID = bigBatch[0].DeviceID
		for i, data := range bigBatch {
			deviceIDs[i] = data.DeviceID
			timestamps[i] = data.Timestamp
			if data.Timestamp < minTimestamp {
				minTimestamp = data.Timestamp
			}
			if data.Timestamp > maxTimestamp {
				maxTimestamp = data.Timestamp
			}
			if data.DeviceID < minDeviceID {
				minDeviceID = data.DeviceID
			}
			if data.DeviceID > maxDeviceID {
				maxDeviceID = data.DeviceID
			}
		}
		// Sort deviceIDs and timestamps in ascending order
		sort.Strings(deviceIDs) // ascending order for deviceIDs
		sort.Slice(timestamps, func(i, j int) bool {
			return timestamps[i] < timestamps[j] // ascending order for timestamps
		})
		querySets := generateQuery(deviceIDs, timestamps, minTimestamp, maxTimestamp, minDeviceID, maxDeviceID)

		// TODO2:使用Query结合UpdateEdgeWeight更新图的权重
		p.updateGraphByQuery(bigBatch, querySets)
		// TODO3:使用谱聚类算法对图进行聚类，得到类别结果
		clusters := p.spectralClustering(classNumber)
		// TODO4:根据分类结果对bigBatch进行划分，得到batches
		// n是节点数量，clusters中保存了每个节点id对应的类别
		// deviceID ---p.nodeMap---> nodeId -> classId
		// 先将bigBatch中的数据按照deviceID进行排序，然后将bigBatch中的数据按照时间戳进行排序
		// 然后根据时间戳的差值来划分batch
		// Create a map from deviceID to cluster
		deviceClusterMap := make(map[string]int)
		for deviceID, nodeID := range p.nodeMap {
			deviceClusterMap[deviceID] = clusters[nodeID-1]
		}

		// Sort bigBatch by cluster first, then by timestamp
		sort.Slice(bigBatch, func(i, j int) bool {
			if deviceClusterMap[bigBatch[i].DeviceID] != deviceClusterMap[bigBatch[j].DeviceID] {
				return deviceClusterMap[bigBatch[i].DeviceID] < deviceClusterMap[bigBatch[j].DeviceID]
			}
			return bigBatch[i].Timestamp < bigBatch[j].Timestamp
		})

		// Distribute data into batches
		batchIndex := 0
		for i := 0; i < len(bigBatch); i++ {
			batches[batchIndex] = append(batches[batchIndex], bigBatch[i])
			if len(batches[batchIndex]) == 20 {
				batchIndex++
			}
		}
	}
	return batches
}

func (p *Processor)updateGraphByQuery(bigBatch []models.SensorData, querySets [][4]interface{}) {
	for _, query := range querySets {
		startDeviceID := query[0].(string)
		endDeviceID := query[1].(string)
		startTime := query[2].(int64)
		endTime := query[3].(int64)

		// Find all data points that fall within this query's range
		matchingData := make(map[string]bool)
		for _, data := range bigBatch {
			if data.DeviceID >= startDeviceID && data.DeviceID <= endDeviceID &&
				data.Timestamp >= startTime && data.Timestamp <= endTime {
				matchingData[data.DeviceID] = true
			}
		}

		// For each pair of devices that co-occur in this query, increase their edge weight
		matchingDevices := make([]string, 0, len(matchingData))
		for deviceID := range matchingData {
			matchingDevices = append(matchingDevices, deviceID)
		}

		const weightIncrement = 0.1
		for i := 0; i < len(matchingDevices); i++ {
			for j := i + 1; j < len(matchingDevices); j++ {
				currentWeight, _ := p.GetEdgeWeight(matchingDevices[i], matchingDevices[j])
				p.UpdateEdgeWeight(matchingDevices[i], matchingDevices[j], currentWeight+weightIncrement)
			}
		}
	}
}

func (p *Processor) spectralClustering(classNumber int) []int {
	// 执行谱聚类算法
	laplacian := p.GetLaplacianMatrix()
	n, _ := laplacian.Dims()

	// 计算特征值和特征向量
	var eigSym mat.EigenSym
	eigSym.Factorize(laplacian, true)

	// 获取特征值和特征向量
	eigenVals := make([]float64, n)
	eigSym.Values(eigenVals)
	eigenVecs := mat.NewDense(n, n, nil)
	eigSym.VectorsTo(eigenVecs)

	// 提取前k个最小特征值对应的特征向量
	k := classNumber
	features := mat.NewDense(n, k, nil)
	for i := 0; i < n; i++ {
		for j := 0; j < k; j++ {
			features.Set(i, j, eigenVecs.At(i, j))
		}
	}

	// 对特征进行K-means聚类
	clusters := make([]int, n)
	centroids := make([][]float64, k)
	for i := range centroids {
		centroids[i] = make([]float64, k)
		for j := range centroids[i] {
			centroids[i][j] = rand.Float64()
		}
	}

	// 迭代进行K-means聚类
	for iter := 0; iter < 100; iter++ {
		// 分配点到最近的中心
		changed := false
		for i := 0; i < n; i++ {
			minDist := math.MaxFloat64
			bestCluster := 0
			for j := 0; j < k; j++ {
				dist := 0.0
				for d := 0; d < k; d++ {
					diff := features.At(i, d) - centroids[j][d]
					dist += diff * diff
				}
				if dist < minDist {
					minDist = dist
					bestCluster = j
				}
			}
			if clusters[i] != bestCluster {
				changed = true
				clusters[i] = bestCluster
			}
		}

		// 更新中心点
		counts := make([]int, k)
		newCentroids := make([][]float64, k)
		for i := range newCentroids {
			newCentroids[i] = make([]float64, k)
		}
		for i := 0; i < n; i++ {
			c := clusters[i]
			counts[c]++
			for d := 0; d < k; d++ {
				newCentroids[c][d] += features.At(i, d)
			}
		}
		for i := 0; i < k; i++ {
			if counts[i] > 0 {
				for d := 0; d < k; d++ {
					centroids[i][d] = newCentroids[i][d] / float64(counts[i])
				}
			}
		}

		if !changed {
			break
		}
	}
	return clusters
}

func generateQuery(deviceIDs []string, timestamps []int64, minTimestamp int64, maxTimestamp int64, minDeviceID string, maxDeviceID string) [][4]interface{} {
	// Generate approximately 25% of total possible queries
	queryCount := int(float64(len(deviceIDs) * len(timestamps)) * 0.25)
	querySets := make([][4]interface{}, queryCount) // [startDeviceID, endDeviceID, startTime, endTime]

	// Calculate means for both dimensions
	timestampMean := (maxTimestamp + minTimestamp) / 2
	timestampRange := maxTimestamp - minTimestamp
	timestampStdDev := float64(timestampRange) / 6.0

	deviceIDLen := len(deviceIDs)
	deviceIDMean := deviceIDLen / 2
	deviceIDStdDev := float64(deviceIDLen) / 6.0

	// Average square size (as percentage of total range)
	avgSquareSize := 0.1 // 10% of total range

	for i := range querySets {
		// Generate center point of the square using Gaussian distribution
		u1, u2 := rand.Float64(), rand.Float64()
		z1 := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
		centerTimestamp := int64(float64(timestampMean) + z1*timestampStdDev)

		u1, u2 = rand.Float64(), rand.Float64()
		z2 := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
		centerDeviceIndex := int(float64(deviceIDMean) + z2*deviceIDStdDev)

		// Generate random square size (varying between 5% and 15% of total range)
		squareSize := avgSquareSize + (rand.Float64()-0.5)*0.1

		// Calculate square boundaries
		halfTimeRange := int64(float64(timestampRange) * squareSize / 2)
		halfDeviceRange := int(float64(deviceIDLen) * squareSize / 2)

		// Calculate time range
		startTime := centerTimestamp - halfTimeRange
		endTime := centerTimestamp + halfTimeRange

		// Calculate device range
		startDeviceIdx := centerDeviceIndex - halfDeviceRange
		endDeviceIdx := centerDeviceIndex + halfDeviceRange

		// Ensure boundaries are within valid ranges
		if startTime < minTimestamp {
			startTime = minTimestamp
		}
		if endTime > maxTimestamp {
			endTime = maxTimestamp
		}
		if startDeviceIdx < 0 {
			startDeviceIdx = 0
		}
		if endDeviceIdx >= deviceIDLen {
			endDeviceIdx = deviceIDLen - 1
		}

		querySets[i] = [4]interface{}{
			deviceIDs[startDeviceIdx],
			deviceIDs[endDeviceIdx],
			startTime,
			endTime,
		}
	}

	return querySets
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
func (p *Processor) GetLaplacianMatrix() *mat.SymDense {
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

	// Convert to symmetric matrix
	symLaplacian := mat.NewSymDense(n, nil)
	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			symLaplacian.SetSym(i, j, laplacian.At(i, j))
		}
	}
	return symLaplacian
}


