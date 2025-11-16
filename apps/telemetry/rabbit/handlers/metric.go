package handlers

import (
	"context"
	"encoding/json"
	"log"
	"telemetry/db"
	"telemetry/models"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type tmpMsg struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorId    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

type DeviceMetricHandler struct {
	DB *db.DB
}

func (h *DeviceMetricHandler) Handle(msg amqp091.Delivery) error {
	log.Printf("Received message: %s\n", string(msg.Body))

	var data models.TelemetryData

	err := json.Unmarshal(msg.Body, &data)
	if err != nil {
		return err
	}

	err = h.DB.AddTemperature(context.TODO(), data)
	if err != nil {
		return err
	}

	if err = msg.Ack(false); err != nil {
		return err
	}

	log.Print("Telemetry saved")

	return nil
}
