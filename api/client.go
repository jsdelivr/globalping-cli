package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

const UserAgent = "Globalping API Go Client"

const (
	posturl   = "https://api.globalping.io/v1/measurements"
	userAgent = UserAgent + "/ v1" + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"
)

// Post measurement to Globalping API
func PostAPI(measurement PostMeasurement) (PostResponse, error) {
	// Format post data
	postData, err := json.Marshal(measurement)
	if err != nil {
		return PostResponse{}, err
	}
	fmt.Println(string(postData))

	// Create a new request
	req, err := http.NewRequest("POST", posturl, bytes.NewBuffer(postData))
	if err != nil {
		return PostResponse{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return PostResponse{}, err
	}
	defer resp.Body.Close()

	// If an error is returned
	if resp.StatusCode == http.StatusBadRequest {
		// Decode the response body as JSON
		var data PostError
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			fmt.Println(err)
			return PostResponse{}, err
		}

		// Print the unknown JSON keys
		fmt.Print(data)
		for k, v := range data.Error.Params {
			fmt.Println(k, v)
		}
		return PostResponse{}, fmt.Errorf("unknown JSON keys")
	}

	// Read the response body
	var data PostResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println(err)
		return PostResponse{}, err
	}

	// Print the struct
	fmt.Println(data)
	return data, nil
}

// Get measurement from Globalping API
