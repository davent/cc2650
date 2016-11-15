package main

import (
	"fmt"
	"github.com/davent/cc2650"
	"log"
)

func main() {

	// Create a new CC2650 SensorTag client
	st, err := cc2650.NewSensorTag()
	if err != nil {
		log.Fatalf("Could not connect to %s\n", cc2650.TAG_NAME)
	}

	// Get remaining battery life percentage
	batt, err := st.Battery.GetRemaing()
	if err != nil {
		log.Fatalf("Could not get battery life info: %s\n", err)
	}
	fmt.Printf("Battery remaining: %d%%\n", batt)

	st.Disconnect()
}
