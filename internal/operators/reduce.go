package operators

import (
	"fmt"
)

// ReduceOperator reduce valores por clave
type ReduceOperator struct {
	name     string
	keyField string
	function string
	params   map[string]interface{}
}

// NewReduceOperator crea un nuevo operador reduce
func NewReduceOperator(keyField, function string, params map[string]interface{}) *ReduceOperator {
	return &ReduceOperator{
		name:     "reduce_by_key",
		keyField: keyField,
		function: function,
		params:   params,
	}
}

// Execute ejecuta el reduce
func (r *ReduceOperator) Execute(input []Record) ([]Record, error) {
	if r.function == "count" {
		return r.count(input)
	}
	
	// TODO: Implementar otras funciones de reduce
	return nil, fmt.Errorf("reduce function %s not yet implemented", r.function)
}

// Name retorna el nombre del operador
func (r *ReduceOperator) Name() string {
	return r.name
}

// count cuenta ocurrencias por clave
func (r *ReduceOperator) count(input []Record) ([]Record, error) {
	counts := make(map[interface{}]int)
	
	for _, record := range input {
		key, exists := record[r.keyField]
		if !exists {
			continue
		}
		counts[key]++
	}
	
	results := make([]Record, 0, len(counts))
	for key, count := range counts {
		results = append(results, Record{
			r.keyField: key,
			"count":    count,
		})
	}
	
	return results, nil
}