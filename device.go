package prolink

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Defined device types.
const (
	DeviceTypeCDJ   DeviceType = 0x01
	DeviceTypeMixer DeviceType = 0x03
	DeviceTypeRB    DeviceType = 0x04
)

var deviceTypeLabels = map[DeviceType]string{
	DeviceTypeCDJ:   "cdj",
	DeviceTypeMixer: "djm",
	DeviceTypeRB:    "rekordbox",
}

// VirtualCDJName is the name given to the Virtual CDJ device.
const VirtualCDJName = "prolink-go"

// VirtualCDJFirmware is a string indicating the firmware version reported with
// status packets.
const VirtualCDJFirmware = "1.43"

// DeviceType represents the types of devices on the network.
type DeviceType byte

// String returns a string representation of a device.
func (d DeviceType) String() string {
	return deviceTypeLabels[d]
}

// DeviceID represents the ID of the device. For CDJs this is the number
// displayed on screen.
type DeviceID byte

// Device represents a device on the network.
type Device struct {
	Name       string
	ID         DeviceID
	Type       DeviceType
	MacAddr    net.HardwareAddr
	IP         net.IP
	LastActive time.Time
}

// String returns a string representation of a device.
func (d *Device) String() string {
	return fmt.Sprintf("%s %02d @ %s [%s]", d.Name, d.ID, d.IP, d.MacAddr)
}

// A DeviceListener responds to devices being added and removed from the PRO DJ
// LINK network.
type DeviceListener interface {
	OnChange(*Device)
}

// The DeviceListenerFunc is an adapter to allow a function to be used as a
// listener for device changes.
type DeviceListenerFunc func(*Device)

// OnChange implements the DeviceListener interface.
func (f DeviceListenerFunc) OnChange(d *Device) { f(d) }

// DeviceManager provides functionality for watching the connection status of
// PRO DJ LINK devices on the network.
type DeviceManager struct {
	delHandlers map[string]DeviceListener
	addHandlers map[string]DeviceListener
	devices     map[DeviceID]*Device
	timers      map[DeviceID]*time.Timer
	allowedNet  *net.IPNet
	lock        sync.RWMutex
}

func (m *DeviceManager) setInterface(iface *net.Interface) error {
	if iface == nil {
		return fmt.Errorf("No network interface provided")
	}
	ipNet, err := getV4IPNetOfInterface(iface)
	if err != nil {
		return err
	}
	if ipNet == nil {
		return fmt.Errorf("No IPv4 address available on interface")
	}
	m.lock.Lock()
	m.allowedNet = &net.IPNet{
		IP:   append(net.IP(nil), ipNet.IP...),
		Mask: append(net.IPMask(nil), ipNet.Mask...),
	}
	m.lock.Unlock()
	return nil
}

// OnDeviceAdded registers a listener that will be called when any PRO DJ LINK
// devices are added to the network. Provide a key if you wish to remove the
// handler later with RemoveListener by specifying the same key.
func (m *DeviceManager) OnDeviceAdded(key string, fn DeviceListener) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.addHandlers[key] = fn
}

// OnDeviceRemoved registers a listener that will be called when any PRO DJ
// LINK devices are removed from the network. Provide a key if you wish to
// remove the handler later with RemoveListener by specifying the same key.
func (m *DeviceManager) OnDeviceRemoved(key string, fn DeviceListener) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.delHandlers[key] = fn
}

// RemoveListener removes a DeviceListener that may have been added by
// OnDeviceAdded or OnDeviceRemoved. Use the key you provided when adding the
// handler.
func (m *DeviceManager) RemoveListener(key string, fn DeviceListener) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.addHandlers, key)
	delete(m.delHandlers, key)
}

// ActiveDeviceMap returns a mapping of device IDs to their associated devices.
func (m *DeviceManager) ActiveDeviceMap() map[DeviceID]*Device {
	m.lock.RLock()
	defer m.lock.RUnlock()
	devices := make(map[DeviceID]*Device, len(m.devices))
	for id, device := range m.devices {
		devices[id] = device
	}
	return devices
}

// ActiveDevices returns a list of active devices on the PRO DJ LINK network.
func (m *DeviceManager) ActiveDevices() []*Device {
	m.lock.RLock()
	defer m.lock.RUnlock()
	devices := make([]*Device, 0, len(m.devices))

	for _, dev := range m.devices {
		devices = append(devices, dev)
	}

	return devices
}

func (m *DeviceManager) expireDevice(dev *Device) {
	m.lock.Lock()
	if m.devices[dev.ID] != dev {
		m.lock.Unlock()
		return
	}
	delete(m.devices, dev.ID)
	delete(m.timers, dev.ID)
	handlers := make([]DeviceListener, 0, len(m.delHandlers))
	for _, handler := range m.delHandlers {
		handlers = append(handlers, handler)
	}
	m.lock.Unlock()

	Log.Info("Device timeout", "device", dev)
	for _, handler := range handlers {
		go handler.OnChange(dev)
	}
}

// activate triggers the DeviceManager to begin watching for device changes on
// the PRO DJ LINK network.
func (m *DeviceManager) activate(announceConn *net.UDPConn) {
	Log.Info("Now monitoring for PROLINK devices")

	announceLock := sync.Mutex{}

	announceHandler := func() {
		packet := make([]byte, announcePacketLen)

		n, source, err := announceConn.ReadFromUDP(packet)
		if err != nil || n != announcePacketLen || source == nil {
			return
		}
		dev, err := deviceFromAnnouncePacket(packet[:n])
		if err != nil {
			return
		}
		if !source.IP.Equal(dev.IP) || source.IP.IsLoopback() ||
			source.IP.IsUnspecified() || source.IP.IsMulticast() {
			return
		}
		m.lock.RLock()
		allowed := m.allowedNet == nil || m.allowedNet.Contains(dev.IP)
		m.lock.RUnlock()
		if !allowed {
			return
		}

		if dev.Name == VirtualCDJName {
			return
		}

		m.lock.Lock()
		// Update device keepalive
		if existing, ok := m.devices[dev.ID]; ok {
			timeout, ok := m.timers[dev.ID]
			if !ok {
				m.lock.Unlock()
				return
			}

			timeout.Stop()
			timeout.Reset(deviceTimeout)
			existing.LastActive = time.Now()
			m.lock.Unlock()
			return
		}
		m.lock.Unlock()

		announceLock.Lock()
		defer announceLock.Unlock()

		// New device
		m.lock.Lock()
		if _, exists := m.devices[dev.ID]; exists {
			m.lock.Unlock()
			return
		}
		m.devices[dev.ID] = dev
		m.timers[dev.ID] = time.AfterFunc(deviceTimeout, func() {
			m.expireDevice(dev)
		})
		handlers := make([]DeviceListener, 0, len(m.addHandlers))
		for _, handler := range m.addHandlers {
			handlers = append(handlers, handler)
		}
		m.lock.Unlock()

		Log.Info("New device tracked", "device", dev)

		for _, handler := range handlers {
			go handler.OnChange(dev)
		}
	}

	// Begin listening for announce packets
	go func() {
		for {
			announceHandler()
		}
	}()
}

func newDeviceManager() *DeviceManager {
	return &DeviceManager{
		addHandlers: map[string]DeviceListener{},
		delHandlers: map[string]DeviceListener{},
		devices:     map[DeviceID]*Device{},
		timers:      map[DeviceID]*time.Timer{},
	}
}
