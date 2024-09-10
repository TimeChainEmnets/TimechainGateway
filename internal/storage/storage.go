package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"timechain-gateway/internal/config"
	"timechain-gateway/pkg/models"
)

type Storage struct {
	config *config.Config
}

func NewStorage(cfg *config.Config) *Storage {
	return &Storage{config: cfg}
}

func (s *Storage) SaveRawData(data []models.RawData) error {
	filename := filepath.Join(s.config.StorageConfig.DataDir, fmt.Sprintf("raw_data_%d.json", data[0].Timestamp))
	return s.saveJSON(filename, data)
}

func (s *Storage) SaveProcessedData(data models.ProcessedData) error {
	filename := filepath.Join(s.config.StorageConfig.DataDir, fmt.Sprintf("processed_data_%d.json", data.Batches[0].Timestamp))
	return s.saveJSON(filename, data)
}

func (s *Storage) LoadRawData(timestamp int64) ([]models.RawData, error) {
	filename := filepath.Join(s.config.StorageConfig.DataDir, fmt.Sprintf("raw_data_%d.json", timestamp))
	var data []models.RawData
	err := s.loadJSON(filename, &data)
	return data, err
}

func (s *Storage) LoadProcessedData(timestamp int64) (models.ProcessedData, error) {
	filename := filepath.Join(s.config.StorageConfig.DataDir, fmt.Sprintf("processed_data_%d.json", timestamp))
	var data models.ProcessedData
	err := s.loadJSON(filename, &data)
	return data, err
}

func (s *Storage) saveJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func (s *Storage) loadJSON(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(data)
}
