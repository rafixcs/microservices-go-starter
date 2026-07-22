package main

import (
	"context"
	"log"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type tripEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripConsumer(rmq *messaging.RabbitMQ) *tripEventConsumer {
	return &tripEventConsumer{
		rabbitmq: rmq,
	}
}

func (c *tripEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages("hello", func(ctx context.Context, msg amqp091.Delivery) error {
		log.Printf("driver received message: %s", msg.Body)
		return nil
	})
}
