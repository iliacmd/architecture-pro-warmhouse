package models

import "time"

const (
	InactiveStatus = "inactive"
	ActiveStatus   = "active"

	HTTPDeviceProtocol = "http"

	DeviceStatusTelemetryTypeID = 4
)

type Device struct {
	ID               int       `json:"id"`
	DeviceTypeID     int       `json:"device_type_id"`
	Name             string    `json:"name"`
	Location         string    `json:"location"`
	Status           string    `json:"status"`
	ConnectionStatus string    `json:"connection_status"`
	UpdatedAt        time.Time `json:"updated_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type DeviceType struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Capabilities map[string]any `json:"capabilities"`
	ProtocolType string         `json:"protocol_type"`
}

type Command struct {
	DeviceID       string         `json:"device_id"`
	Type           string         `json:"command_type"`
	CommandPayload map[string]any `json:"command_payload"`
}

type TelemetryData struct {
	Time            time.Time `json:"time"`
	DeviceID        int       `json:"device_id"`
	TelemetryTypeID int       `json:"telemetry_type_id"`
	StringValue     string    `json:"string_value"`
}
