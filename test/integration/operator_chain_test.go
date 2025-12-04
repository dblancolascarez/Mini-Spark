package integration

import (
	"os"
	"testing"

	op "github.com/dblancolascarez/mini-spark/internal/operators"
)

// TestOperatorChain ejecuta read -> flat_map -> reduce -> write (en memoria)
func TestOperatorChain(t *testing.T) {
	// Preparar input simulado
	input := []op.Record{{"text": "hello spark"}, {"text": "hello mini"}}

	// Operadores
	flat := op.NewFlatMapOperator("split_words", map[string]interface{}{})
	red := op.NewReduceOperator("word", "count", map[string]interface{}{})
	writer := op.NewCSVWriter("data/output/test_chain.csv", map[string]interface{}{})

	// Ejecutar pipeline
	words, err := flat.Execute(input)
	if err != nil {
		t.Fatalf("flat_map error: %v", err)
	}
	counts, err := red.Execute(words)
	if err != nil {
		t.Fatalf("reduce error: %v", err)
	}
	_, err = writer.Execute(counts)
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Validar que el archivo existe
	if _, err := os.Stat("data/output/test_chain.csv"); err != nil {
		t.Fatalf("expected output file, got error: %v", err)
	}
}

// Tests de cadenas de operadores
