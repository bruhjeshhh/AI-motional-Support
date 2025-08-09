package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// A struct to hold the JSON data we send to the Python server.
type RequestPayload struct {
	Text string `json:"text"`
}

// A struct to parse the response from the Python server.
type EmotionResponse struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

// getEmotionFromLocalServer sends text to our Python server and returns the detected emotions.
func getEmotionFromLocalServer(text string) ([]EmotionResponse, error) {
	// Our local server address.
	apiURL := "http://127.0.0.1:5000/analyze_emotion"

	// Prepare the request body as JSON.
	requestBody := RequestPayload{Text: text}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create a new HTTP POST request.
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the required header.
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w. Make sure your Python server is running.", err)
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for a successful response code.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	// Unmarshal the JSON response into a slice of slices of our Go struct.
	var result [][]EmotionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Since the outer array only contains one element, we can just return that inner slice.
	if len(result) > 0 {
		return result[0], nil
	}

	return nil, nil
}

// main function to test our emotion detection.
func main() {
	texts := []string{
		"I'm so happy and excited about my new project!",
		"Feeling a little down and sad today.",
		"I can't believe this happened. I'm so angry!",
		"This is a very surprising development.",
	}

	for _, text := range texts {
		emotions, err := getEmotionFromLocalServer(text)
		if err != nil {
			fmt.Printf("Error for text '%s': %v\n", text, err)
			continue
		}

		fmt.Printf("Text: '%s'\n", text)
		fmt.Println("Detected Emotions:")
		for _, emotion := range emotions {
			fmt.Printf("  - %s (%.2f%%)\n", emotion.Label, emotion.Score*100)
		}
		fmt.Println()
	}
}
