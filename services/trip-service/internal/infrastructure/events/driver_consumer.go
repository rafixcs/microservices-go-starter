package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"

	pbd "ride-sharing/shared/proto/driver"
)

type driverEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  service.TripService
}

func NewDriverConsumer(rmq *messaging.RabbitMQ, service *service.TripService) *driverEventConsumer {
	return &driverEventConsumer{
		rabbitmq: rmq,
		service:  *service,
	}
}

func (c *driverEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.DriverTripResponseQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		var message contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return err
		}

		var payload messaging.DriverTripResponseData
		if err := json.Unmarshal(message.Data, &payload); err != nil {
			log.Printf("failed to unmarshal trip data: %v", err)
			return err
		}

		log.Printf("driver response message: %v", payload)

		switch msg.RoutingKey {
		case contracts.DriverCmdTripAccept:
			if err := c.handleTripAccepted(ctx, payload.TripID, payload.Driver); err != nil {
				log.Printf("Failed to handle the trip accept: %v", err)
				return err
			}
		case contracts.DriverCmdTripDecline:
			log.Println("Declined")
			return nil
		}

		log.Printf("Unknown trip event: %+v", payload)

		return nil
	})
}

func (c *driverEventConsumer) handleTripAccepted(ctx context.Context, tripID string, driver *pbd.Driver) error {
	// Fetch the first
	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	if trip == nil {
		return fmt.Errorf("Trip was not found %s", tripID)
	}

	// Update the trip
	if err := c.service.UpdateTrip(ctx, tripID, "accepted", driver); err != nil {
		log.Printf("Failed to update the trip: %v", err)
		return err
	}

	trip, err = c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	// Driver has been assigned -> pub this event to RabbitMQ
	marshalledTrip, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverAssigned, &contracts.AmqpMessage{
		OwnerID: trip.UserId,
		Data:    marshalledTrip,
	}); err != nil {
		log.Printf("Failed to publish trip driver assigned: %v", err)
		return err
	}

	// TODO: Notify the payment service to start a payment link

	return nil
}
