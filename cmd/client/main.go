// CLI para interactuar con el cluster
package main

import (
	"fmt"
	"os"
	
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <command> [args]")
		fmt.Println("Commands:")
		fmt.Println("  submit <job.json>  - Submit a new job")
		fmt.Println("  status <job-id>    - Get job status")
		fmt.Println("  results <job-id>   - Get job results")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "submit":
		fmt.Println("Submitting job...")
		// TODO: Implementar submit
	case "status":
		fmt.Println("Getting status...")
		// TODO: Implementar status
	case "results":
		fmt.Println("Getting results...")
		// TODO: Implementar results
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
