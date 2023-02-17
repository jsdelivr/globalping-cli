package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"

	"github.com/jsdelivr/globalping-cli/model"
)

const userAgent = "Globalping API Go Client / v1" + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"

var ApiUrl = "https://api.globalping.io/v1/measurements"

// Post measurement to Globalping API
func PostAPI(measurement model.PostMeasurement) (model.PostResponse, error) {
	// Format post data
	postData, err := json.Marshal(measurement)
	if err != nil {
		return model.PostResponse{}, errors.New("failed to marshal post data")
	}

	// Create a new request
	req, err := http.NewRequest("POST", ApiUrl, bytes.NewBuffer(postData))
	if err != nil {
		return model.PostResponse{}, errors.New("failed to create request")
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.PostResponse{}, errors.New("request failed")
	}
	defer resp.Body.Close()

	// If an error is returned
	if resp.StatusCode != http.StatusAccepted {
		// Decode the response body as JSON
		var data model.PostError

		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			return model.PostResponse{}, errors.New("invalid error format returned")
		}

		// 422 error
		if data.Error.Type == "no_probes_found" {
			return model.PostResponse{}, errors.New("no suitable probes found")
		}

		// 400 error
		if data.Error.Type == "validation_error" {
			for _, v := range data.Error.Params {
				fmt.Printf("err: %s\n", v)
			}
			return model.PostResponse{}, fmt.Errorf("validation error")
		}

		// 500 error
		if data.Error.Type == "api_error" {
			return model.PostResponse{}, errors.New("internal server error - please try again later")
		}

		// If the error type is unknown
		return model.PostResponse{}, fmt.Errorf("unknown error response: %s", data.Error.Type)
	}

	// Read the response body
	var data model.PostResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println(err)
		return model.PostResponse{}, errors.New("invalid post measurement format returned")
	}

	return data, nil
}

// Get measurement from Globalping API
func GetAPI(id string) (model.GetMeasurement, error) {
	// Create a new request
	req, err := http.NewRequest("GET", ApiUrl+"/"+id, nil)
	if err != nil {
		return model.GetMeasurement{}, errors.New("failed to create request")
	}
	req.Header.Set("User-Agent", userAgent)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.GetMeasurement{}, errors.New("request failed")
	}
	defer resp.Body.Close()

	// 404 not found
	if resp.StatusCode == http.StatusNotFound {
		return model.GetMeasurement{}, errors.New("measurement not found")
	}

	// 500 error
	if resp.StatusCode == http.StatusInternalServerError {
		return model.GetMeasurement{}, errors.New("internal server error - please try again later")
	}

	// Read the response body
	var data model.GetMeasurement
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println(err)
		return model.GetMeasurement{}, errors.New("invalid get measurement format returned")
	}

	return data, nil
}

func GetApiJson(id string) (string, error) {
	// Create a new request
	req, err := http.NewRequest("GET", ApiUrl+"/"+id, nil)
	if err != nil {
		return "", errors.New("failed to create request")
	}
	req.Header.Set("User-Agent", userAgent)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("request failed")
	}
	defer resp.Body.Close()

	// 404 not found
	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New("measurement not found")
	}

	// 500 error
	if resp.StatusCode == http.StatusInternalServerError {
		return "", errors.New("internal server error - please try again later")
	}

	// Read the response body
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("failed to read response body")
	}
	respString := string(respBytes)

	return respString, nil
}
