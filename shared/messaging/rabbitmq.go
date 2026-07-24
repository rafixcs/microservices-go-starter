package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/retry"
	"ride-sharing/shared/tracing"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange       = "trip"
	DeadLetterExchange = "dlx"
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

func (r *RabbitMQ) setupDeadLetterExchange() error {
	// Declare the dead letter exchange
	err := r.Channel.ExchangeDeclare(
		DeadLetterExchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter exchange: %v", err)
	}

	// Declare the dead letter queue
	q, err := r.Channel.QueueDeclare(
		DeadLetterQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter queue: %v", err)
	}

	// Bind the queue to the exchange with a wildcard routing key
	err = r.Channel.QueueBind(
		q.Name,
		"#", // wildcard routing key to catch all messages
		DeadLetterExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind dead letter queue: %v", err)
	}

	return nil
}

func (r *RabbitMQ) setupExchangesAndQueues() error {
	// First setup the DLQ exchange and queue
	if err := r.setupDeadLetterExchange(); err != nil {
		return err
	}

	err := r.Channel.ExchangeDeclare(
		TripExchange, // name
		"topic",      // type
		true,         // durability
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange: %s: %v", TripExchange, err)
	}

	if err := r.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{
			contracts.TripEventCreated,
			contracts.TripEventDriverNotInterested,
		},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		DriverCmdTripRequestQueue,
		[]string{contracts.DriverCmdTripRequest},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		DriverTripResponseQueue,
		[]string{contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverNoDriversFoundQueue,
		[]string{contracts.TripEventNoDriversFound},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		[]string{contracts.TripEventDriverAssigned},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		PaymentTripResponseQueue,
		[]string{contracts.PaymentCmdCreateSession},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSessionCreatedQueue,
		[]string{contracts.PaymentEventSessionCreated},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSuccessQueue,
		[]string{contracts.PaymentEventSuccess},
		TripExchange,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message *contracts.AmqpMessage) error {
	log.Printf("Publishing message with routing key: %s", routingKey)

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         jsonMsg,
	}

	return tracing.TracedPublisher(ctx, TripExchange, routingKey, msg, r.publish)
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	// Set prefetch count to 1 for fair dispatch
	// This tells RabbitMQ not to give more than one message to a service at a time.
	// The worker will only get the next message after it has acknowleged the previous one.
	err := r.Channel.Qos(
		1,     // prefetchCount: Limit to 1 unacknowleged message per consumer
		0,     // prefetchSizer: No specific limit on message size
		false, // global: Apply prefetchCount to each consumer individually
	)

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

	go func() {
		for msg := range msgs {
			if err := tracing.TracedConsumer(msg, func(ctx context.Context, d amqp.Delivery) error {
				log.Printf("Received a message: %s", msg.Body)

				cfg := retry.DefaultConfig()
				err := retry.WithBackoff(ctx, cfg, func() error {
					return handler(ctx, d)
				})
				if err != nil {
					log.Printf("Message processing failed after %d retries for message ID: %s, err: %v", cfg.MaxRetries, err)

					// Add failure context before sending to the DLQ
					headers := amqp.Table{}
					if d.Headers != nil {
						headers = d.Headers
					}

					headers["x-death-reason"] = err.Error()
					headers["x-origin-exchange"] = d.Exchange
					headers["x-original-routing-key"] = d.RoutingKey
					headers["x-retry-count"] = cfg.MaxRetries
					d.Headers = headers

					// Reject without requeue - message will go to the DLQ
					_ = d.Reject(false)
					return err
				}

				// Only ack if the handler succeeds
				if ackErr := msg.Ack(false); ackErr != nil {
					log.Printf("ERROR: Failed to Ack message: %v. Message body: %s", ackErr, msg.Body)
				}

				return nil

				/*if err := handler(ctx, msg); err != nil {
					log.Fatalf("failed to handle the message: %v", err)
					// Nack the message. Set requeue to false to avoid immediate redelivery loops.
					// Consider a dead-leter exchange (DLQ) or a more sophisticated retry mechanism for
					// production.
					if nackErr := msg.Nack(false, false); nackErr != nil {
						log.Printf("ERROR: Failed to Nack message: %v", nackErr)
					}

					return err
				}*/
			}); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string) error {
	args := amqp.Table{
		"x-dead-letter-exchange": DeadLetterExchange,
	}

	q, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unusedoo
		false,     // exclusive
		false,     // no-wait
		args,      // arguments
	)

	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messageTypes {
		if err := r.Channel.QueueBind(
			q.Name,   // queue name
			msg,      // routing key
			exchange, // exchange
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue to %s: %v", queueName, err)
		}

	}

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

func (r *RabbitMQ) publish(ctx context.Context, exchange, routingKey string, msg amqp.Publishing) error {
	return r.Channel.PublishWithContext(ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		msg,
	)
}
