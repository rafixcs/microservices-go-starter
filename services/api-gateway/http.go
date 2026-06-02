package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/types"
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

		var respBody types.OsrmApiResponse
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

	writeJSON(w, status, response)
}
