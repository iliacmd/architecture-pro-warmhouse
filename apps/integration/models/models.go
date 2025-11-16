package models

import "time"

const (
	TemperatureTelemetryTypeID = 1
)

type Command struct {
	DeviceID       int            `json:"device_id"`
	DeviceProtocol string         `json:"device_protocol"`
	Type           string         `json:"command_type"`
	CommandPayload map[string]any `json:"command_payload"`
}

type TelemetryData struct {
	Time            time.Time `json:"time"`
	DeviceID        int       `json:"device_id"`
	TelemetryTypeID int       `json:"telemetry_type_id"`
	NumericValue    float64   `json:"numeric_value"`
}
