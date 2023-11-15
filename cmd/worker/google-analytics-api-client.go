package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GoogleAnalyticsAPIClient struct {
	client   http.Client
	apiToken string
}

func NewGoogleAnalyticsAPIClient(
	apiToken string,
) *GoogleAnalyticsAPIClient {
	return &GoogleAnalyticsAPIClient{
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		apiToken: apiToken,
	}
}

type Event struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

type eventRequest struct {
	ClientID string  `json:"client_id"`
	Events   []Event `json:"events"`
}

func (c *GoogleAnalyticsAPIClient) pushEvent(measurementID string, clientID string, events []Event) error {
	eventReq := eventRequest{
		ClientID: clientID,
		Events:   events,
	}

	rawEvent, err := json.Marshal(eventReq)
	if err != nil {
		return err
	}

	reqBody := bytes.NewReader(rawEvent)

	url := fmt.Sprintf("https://www.google-analytics.com/mp/collect?measurement_id=%s&api_secret=%s", measurementID, c.apiToken)
	request, err := http.NewRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return err
	}

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("failed to push events: %s", string(responseBody))
	}

	return nil
}
