package cc2650

import (
	"errors"
	"fmt"
	"log"

	"github.com/currantlabs/gatt"
)

const (
	BATTERY_UUID      string = "180f"
	BATTERY_DATA_UUID string = "2a19"
)

type Battery struct {
	Peripheral          gatt.Peripheral
	Service             *gatt.Service
	DataCharacteristic  *gatt.Characteristic
	PercentageReamining int
}

func NewBattery(p gatt.Peripheral, s *gatt.Service) (*Battery, error) {

	b := &Battery{
		Peripheral: p,
		Service:    s,
	}

	// Get Characteristics
	data_uuid, err := gatt.ParseUUID(BATTERY_DATA_UUID)
	if err != nil {
		return nil, err
	}

	cs, err := b.Peripheral.DiscoverCharacteristics([]gatt.UUID{data_uuid}, b.Service)
	if err != nil {
		return nil, err
	}

	for _, c := range cs {
		switch c.UUID().String() {
		case BATTERY_DATA_UUID:
			b.DataCharacteristic = c
		default:
			return nil, errors.New(fmt.Sprintf("Unknown Characteristic UUID: %s\n", c.UUID().String()))
		}
	}

	return b, nil
}

func (b *Battery) GetValue() ([]byte, error) {

	log.Println("Getting data value from battery monitor")
	data, err := b.Peripheral.ReadCharacteristic(b.DataCharacteristic)
	if err != nil {
		return nil, err
	}
	return data, nil

}

func (b *Battery) ParseData(data []byte) int {

	percentage_remaining := int(data[0])

	return percentage_remaining
}

func (b *Battery) GetRemaing() (int, error) {

	data, err := b.GetValue()
	if err != nil {
		return 0, err
	}

	percentage_remaining := b.ParseData(data)

	return percentage_remaining, nil
}
