package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"time"
)

// HashData 计算给定数据的 SHA256 哈希
func HashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// TimeNow 返回当前的 Unix 时间戳（秒）
func TimeNow() int64 {
	return time.Now().Unix()
}

// CalculateDistance 计算两个点之间的欧几里得距离
func CalculateDistance(point1, point2 []float64) float64 {
	if len(point1) != len(point2) {
		return -1 // 错误情况
	}

	var sum float64
	for i := range point1 {
		diff := point1[i] - point2[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// MinMax 返回一个 float64 切片的最小值和最大值
func MinMax(values []float64) (min, max float64) {
	if len(values) == 0 {
		return 0, 0
	}

	min, max = values[0], values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// Contains 检查一个字符串切片是否包含特定字符串
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
