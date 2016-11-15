package main

import (
	"fmt"
	"github.com/davent/cc2650"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {

	// Catch ctrl-c to close gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Create a new CC2650 SensorTag client
	st, err := cc2650.NewSensorTag()
	if err != nil {
		log.Fatalf("Could not connect to %s: %s\n", cc2650.TAG_NAME, err)
	}

	// Enable the TMP007 sensor (IR Temperature)
	err = st.Temperature.Enabled(true)
	if err != nil {
		log.Fatalln("Could not enable the TMP007 sensor (IR Temperature)")
	}

	// Wait a second for the sensor to take a reading
	log.Printf("Waiting for %d seconds while the sensor takes some readings...", 2)
	time.Sleep(2 * time.Second)

	// Read ambient and IR temperatures from the TMP007 sensor
	temps, err := st.Temperature.GetTemperatures()
	if err != nil {
		log.Fatalf("Could not read data value from the TMP007 sensor (IR Temperature): %s\n", err)
	}
	fmt.Printf("Ambient: %0.2fC, IR: %0.2fC\n", temps.Ambient.Celsius, temps.IR.Celsius)

	/* OR */

	// Enable notifications for the TMP007 sensor (IR Temperature)
	temp_chan, err := st.Temperature.Notifications(true)
	if err != nil {
		log.Fatalf("Could not enable notifications for the TMP007 sensor (IR Temperature): %s\n", err)
	}

Loop:
	for {
		select {
		case temps := <-*temp_chan:
			fmt.Printf("%+v) Ambient: %0.2fC, IR: %0.2fC\n", temps.Timestamp, temps.Ambient.Celsius, temps.IR.Celsius)
		case <-c:
			st.Disconnect()
			break Loop
		}
	}

}
