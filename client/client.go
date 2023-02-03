package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/jsdelivr/globalping-cli/model"
)

const UserAgent = "Globalping API Go Client"

const (
	posturl   = "https://api.globalping.io/v1/measurements"
	userAgent = UserAgent + "/ v1" + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"
)

// Post measurement to Globalping API
func PostAPI(measurement model.PostMeasurement) (model.PostResponse, error) {
	// Format post data
	postData, err := json.Marshal(measurement)
	if err != nil {
		return model.PostResponse{}, err
	}
	fmt.Println(string(postData))

	// Create a new request
	req, err := http.NewRequest("POST", posturl, bytes.NewBuffer(postData))
	if err != nil {
		return model.PostResponse{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.PostResponse{}, err
	}
	defer resp.Body.Close()

	// If an error is returned
	if resp.StatusCode == http.StatusBadRequest {
		// Decode the response body as JSON
		var data model.PostError
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			fmt.Println(err)
			return model.PostResponse{}, err
		}

		// Print the unknown JSON keys
		fmt.Print(data)
		for k, v := range data.Error.Params {
			fmt.Println(k, v)
		}
		return model.PostResponse{}, fmt.Errorf("unknown JSON keys")
	}

	// Read the response body
	var data model.PostResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println(err)
		return model.PostResponse{}, err
	}

	// Print the struct
	fmt.Println(data)
	return data, nil
}

// Get measurement from Globalping API
func GetAPI(id string) (model.GetMeasurement, error) {
	// Create a new request
	req, err := http.NewRequest("GET", posturl+"/"+id, nil)
	if err != nil {
		return model.GetMeasurement{}, err
	}
	req.Header.Set("User-Agent", userAgent)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return model.GetMeasurement{}, err
	}
	defer resp.Body.Close()

	// Read the response body
	var data model.GetMeasurement
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println(err)
		return model.GetMeasurement{}, err
	}

	// Print the struct
	return data, nil
}
