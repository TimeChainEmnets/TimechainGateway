package device

import (
	mqtt "github.com/mochi-mqtt/server/v2"
	"log"
	"sync"
	"timechain-gateway/internal/config"
	"timechain-gateway/pkg/models"
)

type Manager struct {
	devices map[string]models.Device
	mutex   sync.RWMutex
	cfg     *config.Config
	server  *mqtt.Server
}

func NewManager(cfg *config.Config, server *mqtt.Server) *Manager {
	m := &Manager{
		devices: make(map[string]models.Device),
		cfg:     cfg,
		server:  server,
	}
	return m
}

func (m *Manager) GetDevices() []models.Device {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	devices := make([]models.Device, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}
	return devices
}

func (m *Manager) AddDevice(clientID string, deviceType string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.devices[clientID] = models.Device{ID: clientID, Type: deviceType}
}

func (m *Manager) RemoveDevice(clientID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.devices, clientID)
}

func (m *Manager) UpdateDeviceType(clientID string, deviceType string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if device, ok := m.devices[clientID]; ok {
		device.Type = deviceType
		m.devices[clientID] = device
		log.Printf("Updated device type: %s (Type: %s)", clientID, deviceType)
	}
}
