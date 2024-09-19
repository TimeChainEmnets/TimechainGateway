package device

import (
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
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
	m.setupServerHooks()
	return m
}

func (m *Manager) setupServerHooks() {
	m.server.Events.OnConnect = func(cl *mqtt.Client, pk packets.Packet) {
		m.AddDevice(cl.ID, "unknown") // 类型初始设为 unknown
		log.Printf("New device connected: %s", cl.ID)
	}

	m.server.Events.OnDisconnect = func(cl *mqtt.Client, err error) {
		m.RemoveDevice(cl.ID)
		log.Printf("Device disconnected: %s", cl.ID)
	}

	m.server.Events.OnMessage = func(cl *mqtt.Client, pk packets.Packet) {
		// 处理设备发送的消息，可能包含设备类型信息
		if pk.TopicName == m.cfg.MQTT.DeviceInfoTopic {
			// 假设消息内容是设备类型
			m.UpdateDeviceType(cl.ID, string(pk.Payload))
		}
	}
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
