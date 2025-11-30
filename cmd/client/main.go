package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	masterAddr := os.Getenv("MASTER_ADDRESS")
	if masterAddr == "" {
		masterAddr = "localhost:8080"
	}

	switch command {
	case "submit":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client submit <job.json>")
			os.Exit(1)
		}
		submitJob(masterAddr, os.Args[2])
		
	case "status":
		if len(os.Args) < 3 {
			fmt.Println("Usage: client status <job-id>")
			os.Exit(1)
		}
		getStatus(masterAddr, os.Args[2])
		
	case "list":
		listJobs(masterAddr)
		
	case "workers":
		listWorkers(masterAddr)
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Mini-Spark Client CLI")
	fmt.Println()
	fmt.Println("Usage: client mmand> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  submit <job.json>  - Submit a new job")
	fmt.Println("  status <job-id>    - Get job status")
	fmt.Println("  list               - List all jobs")
	fmt.Println("  workers            - List active workers")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  MASTER_ADDRESS     - Master address (default: localhost:8080)")
}

func submitJob(masterAddr, jobFile string) {
	// Leer archivo de job
	data, err := os.ReadFile(jobFile)
	if err != nil {
		fmt.Printf("Error reading job file: %v\n", err)
		os.Exit(1)
	}

	// Enviar al master
	url := fmt.Sprintf("http://%s/api/v1/jobs", masterAddr)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error submitting job: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		fmt.Printf("Job submission failed (status %d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	fmt.Printf("  Job submitted successfully\n")
	fmt.Printf("  Job ID: %s\n", result["job_id"])
	fmt.Printf("  Status: %s\n", result["status"])
}

func getStatus(masterAddr, jobID string) {
	url := fmt.Sprintf("http://%s/api/v1/jobs/%s", masterAddr, jobID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error getting status: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error (status %d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var status map[string]interface{}
	json.Unmarshal(body, &status)
	
	fmt.Printf("Job Status:\n")
	fmt.Printf("  ID: %s\n", status["job_id"])
	fmt.Printf("  Name: %s\n", status["name"])
	fmt.Printf("  Status: %s\n", status["status"])
	fmt.Printf("  Progress: %.1f%%\n", status["progress"])
	fmt.Printf("  Tasks: %v completed / %v total\n", 
		status["completed_tasks"], status["total_tasks"])
	
	if status["started_at"] != nil {
		fmt.Printf("  Started: %s\n", status["started_at"])
	}
	if status["completed_at"] != nil {
		fmt.Printf("  Completed: %s\n", status["completed_at"])
	}
}

func listJobs(masterAddr string) {
	url := fmt.Sprintf("http://%s/api/v1/jobs", masterAddr)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error listing jobs: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	jobs := result["jobs"].([]interface{})
	
	fmt.Printf("Jobs (%d total):\n", len(jobs))
	for _, j := range jobs {
		job := j.(map[string]interface{})
		fmt.Printf("  [%s] %s - %s (%.1f%%)\n", 
			job["job_id"], job["name"], job["status"], job["progress"])
	}
}

func listWorkers(masterAddr string) {
	url := fmt.Sprintf("http://%s/api/v1/workers", masterAddr)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error listing workers: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	workers := result["workers"].([]interface{})
	stats := result["stats"].(map[string]interface{})
	
	fmt.Printf("Workers (%v active / %v total):\n", 
		stats["active_workers"], stats["total_workers"])
	
	for _, w := range workers {
		worker := w.(map[string]interface{})
		fmt.Printf("  [%s] %s - %s (tasks: %v)\n", 
			worker["ID"], worker["Status"], worker["Host"], worker["ActiveTasks"])
	}
}
