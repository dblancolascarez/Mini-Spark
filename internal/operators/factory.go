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

	default:
		return nil, fmt.Errorf("unknown operator: %s", node.Operator)
	}
}
