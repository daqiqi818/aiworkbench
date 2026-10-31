// Package openmontage is the private BeefTV-to-OpenMontage runtime boundary.
package openmontage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"infinite-canvas/backend/internal/dramaedit"
)

type Client struct {
	baseURL, token, brokerURL string
	http                      *http.Client
}

func NewFromEnv() *Client {
	return New(os.Getenv("OPENMONTAGE_RUNNER_URL"), os.Getenv("OPENMONTAGE_RUNNER_TOKEN"), os.Getenv("OPENMONTAGE_SOURCE_BROKER_URL"))
}

func New(baseURL, token, brokerURL string) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, brokerURL: strings.TrimRight(brokerURL, "/"), http: &http.Client{Timeout: 90 * time.Second}}
}

func (c *Client) enabled() error {
	if c.baseURL == "" || c.token == "" || c.brokerURL == "" {
		return fmt.Errorf("OpenMontage 运行器未配置")
	}
	return nil
}

func (c *Client) Create(userID, productionID, title string, preferences dramaedit.DeliveryPreferences, sourceResourceIDs []string, accessToken string) (dramaedit.NativeState, error) {
	if err := c.enabled(); err != nil {
		return dramaedit.NativeState{}, err
	}
	payload := map[string]any{"productionId": productionID, "userId": userID, "title": title, "runtimeVersion": "openmontage-9590f5639a0f", "budgetUsd": 0, "targetDurationSeconds": preferences.MaxDurationSeconds, "platform": preferences.Platform, "sourceResourceIds": sourceResourceIDs, "sourceBrokerURL": c.brokerURL + "/api/internal/openmontage/productions/" + productionID + "/sources", "sourceAccessToken": accessToken}
	return c.call("POST", "/internal/openmontage/productions", payload)
}
func (c *Client) Run(userID, productionID string) (dramaedit.NativeState, error) {
	return c.call("POST", "/internal/openmontage/productions/"+productionID+"/run", map[string]string{"userId": userID})
}
func (c *Client) Status(userID, productionID string) (dramaedit.NativeState, error) {
	return c.call("GET", "/internal/openmontage/productions/"+productionID+"?user_id="+userID, nil)
}
func (c *Client) ApproveStage(userID, productionID, stage, note string) (dramaedit.NativeState, error) {
	return c.call("POST", "/internal/openmontage/productions/"+productionID+"/stages/"+stage+"/approve", map[string]string{"userId": userID, "note": note})
}
func (c *Client) Cancel(userID, productionID string) (dramaedit.NativeState, error) {
	return c.call("POST", "/internal/openmontage/productions/"+productionID+"/cancel?user_id="+userID, nil)
}
func (c *Client) Artifact(userID, productionID, artifact string) (io.ReadCloser, error) {
	if err := c.enabled(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/internal/openmontage/productions/"+productionID+"/artifacts/"+artifact+"?user_id="+userID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Openmontage-Runner-Token", c.token)
	response, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode/100 != 2 {
		defer response.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("OpenMontage artifact request failed: %s", strings.TrimSpace(string(body)))
	}
	return response.Body, nil
}
func (c *Client) call(method, path string, payload any) (dramaedit.NativeState, error) {
	if err := c.enabled(); err != nil {
		return dramaedit.NativeState{}, err
	}
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return dramaedit.NativeState{}, err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return dramaedit.NativeState{}, err
	}
	req.Header.Set("X-Openmontage-Runner-Token", c.token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := c.http.Do(req)
	if err != nil {
		return dramaedit.NativeState{}, err
	}
	defer response.Body.Close()
	if response.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return dramaedit.NativeState{}, fmt.Errorf("OpenMontage request failed: %s", strings.TrimSpace(string(raw)))
	}
	var result struct {
		ProductionID     string `json:"productionId"`
		Status           string `json:"status"`
		CurrentStage     string `json:"currentStage"`
		CheckpointStatus string `json:"checkpointStatus"`
		RequiresApproval bool   `json:"requiresApproval"`
		Error            string `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return dramaedit.NativeState{}, err
	}
	return dramaedit.NativeState{ProductionID: result.ProductionID, Status: result.Status, CurrentStage: result.CurrentStage, CheckpointStatus: result.CheckpointStatus, RequiresApproval: result.RequiresApproval, Error: result.Error}, nil
}
