package cc2650

import (
	"time"
)

const (
	// SensorTag
	TAG_NAME     string = "CC2650 SensorTag"
	TI_BASE_UUID string = "f000000004514000b000000000000000"

	NOTIFICATION_DATA_BUFFER_LENGTH = 1

	GATT_NOTIFICATIONS_UUID string = "2902"

	DISCOVERY_RESPONSE_TIMEOUT time.Duration = 5
)
