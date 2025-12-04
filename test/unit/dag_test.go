package unit

import (
	"testing"

	"github.com/dblancolascarez/mini-spark/internal/dag"
)

func TestDAGTopologicalSort(t *testing.T) {
	g := dag.NewDAG()
	_ = g.AddNode("read", "read_csv", map[string]interface{}{"path": "data/input/words.csv"})
	_ = g.AddNode("flat", "flat_map", map[string]interface{}{"function": "split_words"})
	_ = g.AddNode("reduce", "reduce_by_key", map[string]interface{}{"key": "word", "fn": "count"})
	_ = g.AddEdge("read", "flat")
	_ = g.AddEdge("flat", "reduce")
	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("toposort error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(order))
	}
}

// Tests de DAG
