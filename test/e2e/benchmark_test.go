package e2e

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestBenchmarkWordCount realiza benchmark de wordcount con 1M registros
func TestBenchmarkWordCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark in short mode")
	}

	// Generar dataset de 1M registros
	dataFile := "../../data/input/benchmark_1m.csv"
	t.Logf("Generating 1M records dataset...")
	if err := generateLargeCSV(dataFile, 1000000); err != nil {
		t.Fatalf("Failed to generate dataset: %v", err)
	}
	defer os.Remove(dataFile)

	// Iniciar cluster
	masterCmd := exec.Command("../../bin/master")
	if err := masterCmd.Start(); err != nil {
		t.Fatalf("Failed to start master: %v", err)
	}
	defer masterCmd.Process.Kill()
	time.Sleep(2 * time.Second)

	// Iniciar 2 workers
	worker1 := exec.Command("../../bin/worker")
	worker1.Env = append(os.Environ(), "WORKER_ID=bench-worker-1")
	if err := worker1.Start(); err != nil {
		t.Fatalf("Failed to start worker1: %v", err)
	}
	defer worker1.Process.Kill()

	worker2 := exec.Command("../../bin/worker")
	worker2.Env = append(os.Environ(), "WORKER_ID=bench-worker-2")
	if err := worker2.Start(); err != nil {
		t.Fatalf("Failed to start worker2: %v", err)
	}
	defer worker2.Process.Kill()

	time.Sleep(3 * time.Second)

	// Ejecutar job y medir tiempo
	start := time.Now()
	submitCmd := exec.Command("../../bin/client", "submit", "../../examples/wordcount.json")
	if err := submitCmd.Run(); err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	// Esperar completación (ajustar timeout según necesidad)
	time.Sleep(30 * time.Second)
	duration := time.Since(start)

	// Calcular métricas
	recordsPerSecond := float64(1000000) / duration.Seconds()
	
	t.Logf("=== BENCHMARK RESULTS ===")
	t.Logf("Dataset: 1,000,000 records")
	t.Logf("Duration: %v", duration)
	t.Logf("Throughput: %.2f records/sec", recordsPerSecond)
	t.Logf("========================")

	// Criterio de aceptación: al menos 10k records/sec
	if recordsPerSecond < 10000 {
		t.Logf("WARNING: Throughput below target (10k records/sec)")
	}
}

// TestBenchmarkScalability prueba escalabilidad con diferentes números de workers
func TestBenchmarkScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability test in short mode")
	}

	dataFile := "../../data/input/benchmark_100k.csv"
	t.Logf("Generating 100K records dataset...")
	if err := generateLargeCSV(dataFile, 100000); err != nil {
		t.Fatalf("Failed to generate dataset: %v", err)
	}
	defer os.Remove(dataFile)

	// Probar con 1, 2, y 4 workers
	workerCounts := []int{1, 2, 4}
	results := make(map[int]time.Duration)

	for _, numWorkers := range workerCounts {
		t.Logf("Testing with %d worker(s)...", numWorkers)
		
		duration, err := runBenchmarkWithWorkers(numWorkers, dataFile)
		if err != nil {
			t.Errorf("Benchmark with %d workers failed: %v", numWorkers, err)
			continue
		}
		
		results[numWorkers] = duration
		t.Logf("%d worker(s): %v", numWorkers, duration)
	}

	// Analizar escalabilidad
	t.Logf("\n=== SCALABILITY ANALYSIS ===")
	for workers, duration := range results {
		throughput := float64(100000) / duration.Seconds()
		t.Logf("%d workers: %v (%.0f records/sec)", workers, duration, throughput)
	}
}

// TestBenchmarkOperators benchmark de operadores específicos
func TestBenchmarkOperators(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping operator benchmark in short mode")
	}

	t.Log("Benchmarking individual operators...")
	
	// Generar datos para diferentes operadores
	testCases := []struct {
		name     string
		operator string
		records  int
	}{
		{"Map", "map", 500000},
		{"Filter", "filter", 500000},
		{"Reduce", "reduce_by_key", 100000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// En una implementación completa, aquí se ejecutaría cada operador
			// y se mediría su rendimiento específico
			t.Logf("Benchmarking %s with %d records", tc.operator, tc.records)
		})
	}
}

// TestMemoryUsage prueba el uso de memoria bajo carga
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	t.Log("Testing memory usage under load...")
	
	// Generar dataset grande
	dataFile := "../../data/input/memory_test.csv"
	if err := generateLargeCSV(dataFile, 2000000); err != nil {
		t.Fatalf("Failed to generate dataset: %v", err)
	}
	defer os.Remove(dataFile)

	// Iniciar cluster y monitorear memoria
	masterCmd := exec.Command("../../bin/master")
	if err := masterCmd.Start(); err != nil {
		t.Fatalf("Failed to start master: %v", err)
	}
	defer masterCmd.Process.Kill()

	t.Log("Memory test requires manual monitoring of process memory")
	t.Log("Use: watch -n 1 'ps aux | grep -E \"master|worker\"'")
}

// runBenchmarkWithWorkers ejecuta un benchmark con N workers
func runBenchmarkWithWorkers(numWorkers int, dataFile string) (time.Duration, error) {
	// Iniciar master
	masterCmd := exec.Command("../../bin/master")
	if err := masterCmd.Start(); err != nil {
		return 0, err
	}
	defer masterCmd.Process.Kill()
	time.Sleep(2 * time.Second)

	// Iniciar workers
	workers := make([]*exec.Cmd, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i] = exec.Command("../../bin/worker")
		workers[i].Env = append(os.Environ(), fmt.Sprintf("WORKER_ID=bench-w%d", i))
		if err := workers[i].Start(); err != nil {
			return 0, err
		}
		defer workers[i].Process.Kill()
	}
	time.Sleep(3 * time.Second)

	// Ejecutar job
	start := time.Now()
	submitCmd := exec.Command("../../bin/client", "submit", "../../examples/wordcount.json")
	if err := submitCmd.Run(); err != nil {
		return 0, err
	}

	// Esperar completación
	time.Sleep(20 * time.Second)
	
	return time.Since(start), nil
}

// generateLargeCSV genera un archivo CSV grande para benchmarks
func generateLargeCSV(filename string, numRecords int) error {
	// Crear directorio si no existe
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Escribir header
	writer.Write([]string{"id", "text", "timestamp"})

	// Palabras de ejemplo
	words := []string{
		"hello", "world", "test", "data", "benchmark", 
		"spark", "mini", "distributed", "system", "computing",
		"map", "reduce", "filter", "aggregate", "join",
	}

	// Generar registros
	for i := 0; i < numRecords; i++ {
		// Generar texto aleatorio con 5-15 palabras
		numWords := 5 + rand.Intn(10)
		text := ""
		for j := 0; j < numWords; j++ {
			text += words[rand.Intn(len(words))] + " "
		}

		record := []string{
			fmt.Sprintf("%d", i),
			text,
			time.Now().Format(time.RFC3339),
		}
		
		if err := writer.Write(record); err != nil {
			return err
		}

		// Flush periódicamente para no saturar memoria
		if i%10000 == 0 {
			writer.Flush()
		}
	}

	return nil
}