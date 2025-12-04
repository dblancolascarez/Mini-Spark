package unit

import (
	"testing"
	"time"

	"github.com/dblancolascarez/mini-spark/internal/master"
	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

func TestSchedulerRoundRobin(t *testing.T) {
	coord := master.NewCoordinator(10 * time.Second)
	_ = coord.RegisterWorker(protocol.WorkerRegisterRequest{WorkerID: "w1", Host: "localhost", Port: 8081, Capacity: 1})
	_ = coord.RegisterWorker(protocol.WorkerRegisterRequest{WorkerID: "w2", Host: "localhost", Port: 8082, Capacity: 1})
	sched := master.NewScheduler(coord, master.PolicyRoundRobin)
	wA, _ := sched.SelectWorker()
	wB, _ := sched.SelectWorker()
	if wA.ID == wB.ID {
		t.Fatalf("round-robin should alternate workers")
	}
}

// Tests del planificador
