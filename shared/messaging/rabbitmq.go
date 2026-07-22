package messaging

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Failed to open channel: %s", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	if err := rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("Failed to setup exchanges and queues: %s", err)
	}

	return rmq, nil
}

func (r *RabbitMQ) setupExchangesAndQueues() error {
	_, err := r.Channel.QueueDeclare(
		"hello", // name
		true,    // durable
		false,   // delete when unusedoo
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message string) error {
	return r.Channel.PublishWithContext(ctx,
		"",      // exchange
		"hello", // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(message),
			DeliveryMode: amqp.Persistent,
		},
	)
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack // auto-ack can cause problems of consuming messages even if the service fails so we handle ack to the queue mannualy
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	if err != nil {
		return err
	}

	ctx := context.Background()

	go func() {
		for msg := range msgs {
			log.Printf("Received a message: %s", msg.Body)

			if err := handler(ctx, msg); err != nil {
				log.Fatalf("failed to handle the message: %v", err)
				// Nack the message. Set requeue to false to avoid immediate redelivery loops.
				// Consider a dead-leter exchange (DLQ) or a more sophisticated retry mechanism for
				// production.
				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("ERROR: Failed to Nack message: %v", nackErr)
				}
			}

			// Only ack if the handler succeeds
			_ = msg.Ack(false)
		}
	}()

	return nil
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
	if r.Channel != nil {
		r.Channel.Close()
	}
}
