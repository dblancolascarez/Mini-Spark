package unit

import (
	"testing"

	op "github.com/dblancolascarez/mini-spark/internal/operators"
)

func TestFlatMapSplitWords(t *testing.T) {
	fm := op.NewFlatMapOperator("split_words", map[string]interface{}{})
	input := []op.Record{{"text": "Hello mini spark"}}
	out, err := fm.Execute(input)
	if err != nil {
		t.Fatalf("flat_map error: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 words, got %d", len(out))
	}
	if out[0]["word"] != "hello" {
		t.Fatalf("expected 'hello', got %v", out[0]["word"])
	}
}

func TestReduceByKeyCount(t *testing.T) {
	r := op.NewReduceOperator("word", "count", map[string]interface{}{})
	input := []op.Record{{"word": "a"}, {"word": "b"}, {"word": "a"}}
	out, err := r.Execute(input)
	if err != nil {
		t.Fatalf("reduce error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(out))
	}
}

// Tests unitarios de operadores
