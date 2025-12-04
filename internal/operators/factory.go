package operators

import (
	"fmt"

	"github.com/dblancolascarez/mini-spark/internal/dag"
)

// Factory crea operadores a partir de nodos DAG
type Factory struct{}

// NewFactory crea una nueva factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateOperator crea un operador basado en un nodo DAG
func (f *Factory) CreateOperator(node *dag.Node) (Operator, error) {
	switch node.Operator {
	case "read_csv":
		path, ok := node.Parameters["path"].(string)
		if !ok {
			return nil, fmt.Errorf("read_csv requires 'path' parameter")
		}
		return NewCSVReader(path, node.Parameters), nil

	case "read_jsonl":
		path, ok := node.Parameters["path"].(string)
		if !ok {
			return nil, fmt.Errorf("read_jsonl requires 'path' parameter")
		}
		return NewJSONLReader(path, node.Parameters), nil

	case "write_csv":
		path, ok := node.Parameters["path"].(string)
		if !ok {
			return nil, fmt.Errorf("write_csv requires 'path' parameter")
		}
		return NewCSVWriter(path, node.Parameters), nil

	case "write_jsonl":
		path, ok := node.Parameters["path"].(string)
		if !ok {
			return nil, fmt.Errorf("write_jsonl requires 'path' parameter")
		}
		return NewJSONLWriter(path, node.Parameters), nil

	case "map":
		function, ok := node.Parameters["function"].(string)
		if !ok {
			return nil, fmt.Errorf("map requires 'function' parameter")
		}
		return NewMapOperator(function, node.Parameters), nil

	case "filter":
		condition, ok := node.Parameters["condition"].(string)
		if !ok {
			return nil, fmt.Errorf("filter requires 'condition' parameter")
		}
		return NewFilterOperator(condition, node.Parameters), nil

	case "flat_map":
		function, ok := node.Parameters["function"].(string)
		if !ok {
			function = "identity"
		}
		return NewFlatMapOperator(function, node.Parameters), nil

	case "reduce_by_key":
		key, ok := node.Parameters["key"].(string)
		if !ok {
			return nil, fmt.Errorf("reduce_by_key requires 'key' parameter")
		}
		// Aceptar 'fn' o 'function' para compatibilidad
		fn, ok := node.Parameters["fn"].(string)
		if !ok {
			fn, ok = node.Parameters["function"].(string)
			if !ok {
				fn = "count" // default
			}
		}
		return NewReduceOperator(key, fn, node.Parameters), nil

	case "join":
		joinType, ok := node.Parameters["type"].(string)
		if !ok {
			joinType = "inner" // default
		}
		leftKey, ok := node.Parameters["left_key"].(string)
		if !ok {
			return nil, fmt.Errorf("join requires 'left_key' parameter")
		}
		rightKey, ok := node.Parameters["right_key"].(string)
		if !ok {
			return nil, fmt.Errorf("join requires 'right_key' parameter")
		}
		return NewJoinOperator(joinType, leftKey, rightKey, node.Parameters), nil

	case "aggregate":
		fn, ok := node.Parameters["fn"].(string)
		if !ok {
			return nil, fmt.Errorf("aggregate requires 'fn' parameter")
		}
		return NewAggregateOperator(fn, node.Parameters), nil

	case "shuffle":
		partitions := 4 // default
		if p, ok := node.Parameters["partitions"].(float64); ok {
			partitions = int(p)
		}
		return NewShuffleOperator(partitions, node.Parameters), nil

	default:
		return nil, fmt.Errorf("unknown operator: %s", node.Operator)
	}
}
