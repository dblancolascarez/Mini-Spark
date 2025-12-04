package e2e

import (
	"os/exec"
	"testing"
	"time"
)

// TestMultinode arranca master y 2 workers y envía un job
func TestMultinode(t *testing.T) {
	masterCmd := exec.Command("../../bin/master")
	if err := masterCmd.Start(); err != nil {
		t.Fatalf("start master: %v", err)
	}
	defer masterCmd.Process.Kill()

	w1 := exec.Command("../../bin/worker")
	w1.Env = append(w1.Env, "WORKER_ID=e2e-w1")
	if err := w1.Start(); err != nil {
		t.Fatalf("start worker1: %v", err)
	}
	defer w1.Process.Kill()

	w2 := exec.Command("../../bin/worker")
	w2.Env = append(w2.Env, "WORKER_ID=e2e-w2")
	if err := w2.Start(); err != nil {
		t.Fatalf("start worker2: %v", err)
	}
	defer w2.Process.Kill()

	time.Sleep(3 * time.Second)

	submit := exec.Command("../../bin/client", "submit", "../../examples/wordcount.json")
	if err := submit.Run(); err != nil {
		t.Fatalf("submit job: %v", err)
	}

	time.Sleep(5 * time.Second)
}

// Tests end-to-end multinodo
