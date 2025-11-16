package services

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type MetricService struct {
	QueueCannel *amqp091.Channel
	QueueName   string
}

func NewMetricService(qc *amqp091.Channel, queueName string) *MetricService {
	return &MetricService{
		QueueCannel: qc,
		QueueName:   queueName,
	}
}

func (s *MetricService) Publish(_ context.Context, data any) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = s.QueueCannel.Publish(
		"",          // exchange
		s.QueueName, // routing key
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        []byte(msg),
		},
	)
	if err != nil {
		return err
	}

	return nil
}
