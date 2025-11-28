package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

// Client maneja la comunicación del worker con el master
type Client struct {
	workerID      string
	masterAddress string
	host          string
	port          int
	httpClient    *http.Client
}

// NewClient crea un nuevo cliente worker
func NewClient(workerID, masterAddress, host string, port int) *Client {
	return &Client{
		workerID:      workerID,
		masterAddress: masterAddress,
		host:          host,
		port:          port,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Register registra el worker con el master
func (c *Client) Register(capacity int) error {
	req := protocol.WorkerRegisterRequest{
		WorkerID:   c.workerID,
		Host:       c.host,
		Port:       c.port,
		Capacity:   capacity,
		RegisterAt: time.Now(),
	}

	url := fmt.Sprintf("http://%s/api/v1/workers/register", c.masterAddress)
	
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var registerResp protocol.WorkerRegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !registerResp.Success {
		return fmt.Errorf("registration failed: %s", registerResp.Message)
	}

	return nil
}

// SendHeartbeat envía un heartbeat al master
func (c *Client) SendHeartbeat(activeTasks int, cpuUsage float64, memoryMB int) error {
	req := protocol.HeartbeatRequest{
		WorkerID:      c.workerID,
		Timestamp:     time.Now(),
		ActiveTasks:   activeTasks,
		CPUUsage:      cpuUsage,
		MemoryUsageMB: memoryMB,
	}

	url := fmt.Sprintf("http://%s/api/v1/workers/heartbeat", c.masterAddress)
	
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed with status: %d", resp.StatusCode)
	}

	return nil
}
