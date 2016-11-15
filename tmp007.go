package cc2650

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/currantlabs/gatt"
)

const (
	TMP007_UUID               string = "f000aa0004514000b000000000000000"
	TMP007_DATA_UUID          string = "f000aa0104514000b000000000000000"
	TMP007_CONFIGURATION_UUID string = "f000aa0204514000b000000000000000"
	TMP007_PERIOD_UUID        string = "f000aa0304514000b000000000000000"
)

type TMP007 struct {
	Peripheral                  gatt.Peripheral
	Service                     *gatt.Service
	DataCharacteristic          *gatt.Characteristic
	ConfigurationCharacteristic *gatt.Characteristic
	PeriodCharacteristic        *gatt.Characteristic
	NotificationDescriptor      *gatt.Descriptor

	TempChan chan *Temperatures
}

type Temperature struct {
	Celsius    float64
	Fahrenheit float64
}

type Temperatures struct {
	Timestamp time.Time
	Ambient   *Temperature
	IR        *Temperature
}

func NewTMP007(p gatt.Peripheral, s *gatt.Service) (*TMP007, error) {

	temp_chan := make(chan *Temperatures)

	t := &TMP007{
		Peripheral: p,
		Service:    s,
		TempChan:   temp_chan,
	}

	// Get Characteristics
	data_uuid, err := gatt.ParseUUID(TMP007_DATA_UUID)
	if err != nil {
		return nil, err
	}

	configuration_uuid, err := gatt.ParseUUID(TMP007_CONFIGURATION_UUID)
	if err != nil {
		return nil, err
	}

	period_uuid, err := gatt.ParseUUID(TMP007_PERIOD_UUID)
	if err != nil {
		return nil, err
	}

	notifications_uuid, err := gatt.ParseUUID(GATT_NOTIFICATIONS_UUID)
	if err != nil {
		return nil, err
	}

	cs, err := t.Peripheral.DiscoverCharacteristics([]gatt.UUID{data_uuid, configuration_uuid, period_uuid}, t.Service)
	if err != nil {
		return nil, err
	}

	for _, c := range cs {
		switch c.UUID().String() {
		case TMP007_DATA_UUID:
			t.DataCharacteristic = c

			// Discovery descriptors
			ds, err := t.Peripheral.DiscoverDescriptors([]gatt.UUID{notifications_uuid}, c)
			if err != nil {
				log.Fatalf("Failed to discover descriptors, err: %s\n", err)
				continue
			}
			t.NotificationDescriptor = ds[0]

		case TMP007_CONFIGURATION_UUID:
			t.ConfigurationCharacteristic = c
		case TMP007_PERIOD_UUID:
			t.PeriodCharacteristic = c
		default:
			return nil, errors.New(fmt.Sprintf("Unknown Characteristic UUID: %s\n", c.UUID().String()))
		}
	}

	return t, nil
}

func (t *TMP007) Enabled(bool) error {

	log.Println("Turning on IR Temperature sensor")
	t.ConfigurationCharacteristic.SetValue([]byte{0x01})
	err := t.Peripheral.WriteCharacteristic(t.ConfigurationCharacteristic, []byte{0x01}, true)
	if err != nil {
		return err
	}
	return nil

}

func (t *TMP007) GetValue() ([]byte, error) {

	log.Println("Getting data value from IR Temperature sensor")
	b, err := t.Peripheral.ReadCharacteristic(t.DataCharacteristic)
	if err != nil {
		return nil, err
	}
	return b, nil

}

func (t *TMP007) ParseTempData(b []byte) *Temperatures {
	ambient_data := binary.LittleEndian.Uint16(b[2:])
	ir_data := binary.LittleEndian.Uint16(b[:2])

	ambient_raw := ambient_data >> 2 & 0x3FFF
	ir_raw := ir_data >> 2 & 0x3FFF

	ambient_c := float64(ambient_raw) * 0.03125
	ir_c := float64(ir_raw) * 0.03125

	temps := &Temperatures{
		Timestamp: time.Now(),
		Ambient: &Temperature{
			Celsius: ambient_c,
		},
		IR: &Temperature{
			Celsius: ir_c,
		},
	}
	return temps
}

func (t *TMP007) GetTemperatures() (*Temperatures, error) {
	data, err := t.GetValue()
	if err != nil {
		return nil, err
	}

	return t.ParseTempData(data), nil
}

func (t *TMP007) Notifications(enabled bool) (*chan *Temperatures, error) {

	log.Println("Setting IR Temperature Notifications")
	t.NotificationDescriptor.SetValue([]byte{0x0001})
	err := t.Peripheral.WriteDescriptor(t.NotificationDescriptor, []byte{0x0001})
	if err != nil {
		return nil, err
	}

	// Subscribe the characteristic, if possible.
	if (t.DataCharacteristic.Properties() & (gatt.CharNotify | gatt.CharIndicate)) != 0 {
		f := func(c *gatt.Characteristic, b []byte, err error) {
			// If the channel is full, pop an element off
			if len(t.TempChan) >= NOTIFICATION_DATA_BUFFER_LENGTH {
				<-t.TempChan
			}
			t.TempChan <- t.ParseTempData(b)
		}
		if err := t.Peripheral.SetNotifyValue(t.DataCharacteristic, f); err != nil {
			return nil, err
		}
	}

	return &t.TempChan, nil

}
