package models

import "time"

type TelemetryData struct {
	Time            time.Time `json:"time"`
	DeviceID        int       `json:"device_id"`
	TelemetryTypeID int       `json:"telemetry_type_id"`
	NumericValue    float64   `json:"numeric_value"`
}
