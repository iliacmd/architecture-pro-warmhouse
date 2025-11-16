package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type IntegrationService struct {
	BaseURL    string
	HTTPClient *http.Client
}

type CreateCommandRequest struct {
	DeviceID       string         `json:"device_id"`
	DeviceProtocol string         `json:"device_protocol"`
	Type           string         `json:"command_type"`
	CommandPayload map[string]any `json:"command_payload"`
}

type CreateCommandResponse struct {
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
	Status      string    `json:"status"`
	SensorID    string    `json:"sensor_id"`
	SensorType  string    `json:"sensor_type"`
	Description string    `json:"description"`
}

func NewIntegrationService(baseURL string) *IntegrationService {
	return &IntegrationService{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *IntegrationService) CreateCommand(req CreateCommandRequest) (*CreateCommandResponse, error) {
	url := fmt.Sprintf("%s/api/v1/commands", s.BaseURL)

	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := s.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var cmdResp CreateCommandResponse
	if err = json.NewDecoder(resp.Body).Decode(&cmdResp); err != nil {
		return nil, fmt.Errorf("error decoding command response: %w", err)
	}

	return &cmdResp, nil
}
