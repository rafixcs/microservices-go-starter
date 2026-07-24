package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	var reqBody previewTripRequest
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Failed to parse json data", http.StatusBadRequest)
		return
	}

	if reqBody.UserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(r.Context(), reqBody.toProto())
	if err != nil {
		log.Printf("[trip-service.handleTripPreview] error: %v", err)
		http.Error(w, "failed to preview trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripPreview}

	writeJSON(w, http.StatusCreated, response)
}

func handleTripStart(w http.ResponseWriter, r *http.Request) {
	var reqBody startTripRequest
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Failed to parse json data", http.StatusBadRequest)
		return
	}

	if reqBody.UserID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}

	if reqBody.RideFareID == "" {
		http.Error(w, "RideFare is required", http.StatusBadRequest)
		return
	}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer tripService.Close()

	tripStart, err := tripService.Client.CreateTrip(r.Context(), reqBody.toProto())
	if err != nil {
		log.Printf("[trip-service.handleTripStart] Error: %v", err)
		http.Error(w, "failed to start trip", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{
		Data: tripStart,
	}

	writeJSON(w, http.StatusCreated, response)
}

func handleDriverRegister(w http.ResponseWriter, r *http.Request) {
	var reqBody driverRegisterRequest
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Failed to parse json data", http.StatusBadRequest)
		return
	}

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer driverService.Close()

	driver, err := driverService.Client.RegisterDriver(r.Context(), reqBody.toProto())
	if err != nil {
		log.Printf("[driver-service.handleDriverRegister] Error: %v", err)
		http.Error(w, "failed to register driver", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{
		Data: driver,
	}

	writeJSON(w, http.StatusCreated, response)
}

func handleDriverUnRegister(w http.ResponseWriter, r *http.Request) {
	var reqBody driverUnRegisterRequest
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Failed to parse json data", http.StatusBadRequest)
		return
	}

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer driverService.Close()

	driver, err := driverService.Client.UnRegisterDriver(r.Context(), reqBody.toProto())
	if err != nil {
		log.Printf("[driver-service.handleDriverRegister] Error: %v", err)
		http.Error(w, "failed to unregister driver", http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{
		Data: driver,
	}

	writeJSON(w, http.StatusOK, response)
}

func handleStripeWebhook(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	webhookKey := env.GetString("STRIPE_WEBHOOK_KEY", "")
	if webhookKey == "" {
		log.Printf("Webhook key is required")
		return
	}

	event, err := webhook.ConstructEventWithOptions(
		body,
		r.Header.Get("Stripe-Signature"),
		webhookKey,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v", err)
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	log.Printf("Received Stripe event: %v", event)

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		payload := messaging.PaymentStatusUpdateData{
			TripID:   session.Metadata["trip_id"],
			UserID:   session.Metadata["user_id"],
			DriverID: session.Metadata["driver_id"],
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshalling payload: %v", err)
			http.Error(w, "Failed to marshal payload", http.StatusInternalServerError)
			return
		}

		message := &contracts.AmqpMessage{
			OwnerID: session.Metadata["user_id"],
			Data:    payloadBytes,
		}

		if err := rb.PublishMessage(
			r.Context(),
			contracts.PaymentEventSuccess,
			message,
		); err != nil {
			log.Printf("Error publishing payment event: %v", err)
			http.Error(w, "Failed to publish payment event", http.StatusInternalServerError)
			return
		}
	}
}

/*
response := contracts.APIResponse{}

	urlPath := tripServiceAddr + "/preview"
	log.Printf("urlPath: %s", urlPath)
	encoded, err := json.Marshal(reqBody)
	if err != nil {
		http.Error(w, "Failed to encode request", http.StatusInternalServerError)
		return
	}
	resp, err := http.Post(urlPath, "application/json", bytes.NewReader(encoded))

	var status int
	if err != nil {
		response.Error = &contracts.APIError{
			Code:    "1234",
			Message: err.Error(),
		}
		status = http.StatusInternalServerError
	} else if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		response.Error = &contracts.APIError{
			Code:    "4321",
			Message: fmt.Sprintf("Trip service returned error [%s]: %s", resp.Status, string(body)),
		}
		status = http.StatusInternalServerError
	} else {
		defer r.Body.Close()

		var respBody tripTypes.OsrmApiResponse
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			response.Error = &contracts.APIError{
				Code:    "43211",
				Message: fmt.Sprintf("failed to parse response body: %w", err),
			}
			status = http.StatusInternalServerError
		}
		log.Printf("trip preview resp: %s", respBody)
		response.Data = respBody
		status = http.StatusCreated
	}
*/
