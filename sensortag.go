package cc2650

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/currantlabs/gatt"
)

type SensorTag struct {
	Device      gatt.Device
	Peripheral  gatt.Peripheral
	Battery     *Battery
	Temperature *TMP007
}

var DefaultClientOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, true),
}

var done = make(chan struct{})
var discovered_device = make(chan struct{})
var discovered_peripheral = make(chan struct{})

func onStateChanged(d gatt.Device, s gatt.State) {
	log.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		log.Printf("Looking for %ss...\n", TAG_NAME)
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	if a.LocalName == TAG_NAME {
		log.Printf("Found %s: %s\n", p.Name(), p.ID())

		p.Device().Connect(p)
	}
}

func (st *SensorTag) onPeriphConnected(p gatt.Peripheral, err error) {
	log.Printf("Connected to %s: %s\n", TAG_NAME, p.ID())
	st.Peripheral = p

	if err := p.SetMTU(500); err != nil {
		log.Fatalf("Failed to set MTU, err: %s\n", err)
	}

	// Discovery services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		log.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	for _, s := range ss {

		switch s.UUID().String() {
		case TMP007_UUID:
			log.Println("Found a tmp007 sensor")
			temp, err := NewTMP007(p, s)
			if err != nil {
				log.Fatal("Could not load tmp007 sensor")
				continue
			}
			st.Temperature = temp
		case BATTERY_UUID:
			log.Println("Found a battery")
			battery, err := NewBattery(p, s)
			if err != nil {
				log.Fatal("Could not load battery monitor")
				continue
			}
			st.Battery = battery
		}

	}
	close(discovered_peripheral)
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	log.Println("Disconnected")
	close(done)
}

func NewSensorTag() (*SensorTag, error) {

	// Create a new SensorTag object
	st := &SensorTag{}

	// Open the BLE device
	device_err_chan := make(chan error)
	go func() {
		d, err := gatt.NewDevice(DefaultClientOptions...)
		if err != nil {
			device_err_chan <- err
		}
		st.Device = d

		close(discovered_device)
	}()

	// Don't wait forever
	select {
	case <-discovered_device:
	case err := <-device_err_chan:
		return nil, err
	case <-time.After(DISCOVERY_RESPONSE_TIMEOUT * time.Second):
		return nil, errors.New(fmt.Sprintf("Could find BLE device within %d seconds", DISCOVERY_RESPONSE_TIMEOUT))
	}

	// Register handlers.
	st.Device.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(st.onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)
	st.Device.Init(onStateChanged)

	// Don't wait forever
	select {
	case <-discovered_peripheral:
	case <-time.After(DISCOVERY_RESPONSE_TIMEOUT * time.Second):
		return nil, errors.New(fmt.Sprintf("Could not connect to a %s within %d seconds", TAG_NAME, DISCOVERY_RESPONSE_TIMEOUT))
	}
	return st, nil
}

func (st *SensorTag) Disconnect() {
	if st.Peripheral != nil {
		st.Peripheral.Device().CancelConnection(st.Peripheral)
	}
	<-done
}
