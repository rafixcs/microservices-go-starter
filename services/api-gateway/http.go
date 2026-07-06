package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
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
