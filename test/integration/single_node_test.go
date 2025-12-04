package integration

import (
	"testing"

	op "github.com/dblancolascarez/mini-spark/internal/operators"
)

// TestSingleNodePipeline ejecuta pipeline en un solo proceso
func TestSingleNodePipeline(t *testing.T) {
	input := []op.Record{{"text": "hello world"}}
	flat := op.NewFlatMapOperator("split_words", map[string]interface{}{})
	out, err := flat.Execute(input)
	if err != nil {
		t.Fatalf("flat_map error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 words, got %d", len(out))
	}
}
