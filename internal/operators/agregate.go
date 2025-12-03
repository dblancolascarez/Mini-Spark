package operators

import (
	"fmt"
)

// AggregateOperator realiza agregaciones sobre datos
type AggregateOperator struct {
	name     string
	function string
	params   map[string]interface{}
}

// NewAggregateOperator crea un nuevo operador aggregate
func NewAggregateOperator(function string, params map[string]interface{}) *AggregateOperator {
	return &AggregateOperator{
		name:     "aggregate",
		function: function,
		params:   params,
	}
}

// Execute ejecuta la agregación
func (a *AggregateOperator) Execute(input []Record) ([]Record, error) {
	if len(input) == 0 {
		return []Record{}, nil
	}
	
	switch a.function {
	case "count":
		return a.count(input)
	case "sum":
		return a.sum(input)
	case "avg", "average":
		return a.avg(input)
	case "max":
		return a.max(input)
	case "min":
		return a.min(input)
	case "group_by":
		return a.groupBy(input)
	default:
		return nil, fmt.Errorf("unsupported aggregate function: %s", a.function)
	}
}

// Name retorna el nombre del operador
func (a *AggregateOperator) Name() string {
	return a.name
}

// count cuenta registros
func (a *AggregateOperator) count(input []Record) ([]Record, error) {
	return []Record{{"count": len(input)}}, nil
}

// sum suma valores de un campo
func (a *AggregateOperator) sum(input []Record) ([]Record, error) {
	field, ok := a.params["field"].(string)
	if !ok {
		return nil, fmt.Errorf("sum requires 'field' parameter")
	}
	
	var sum float64
	for _, record := range input {
		if val, exists := record[field]; exists {
			switch v := val.(type) {
			case int:
				sum += float64(v)
			case int64:
				sum += float64(v)
			case float64:
				sum += v
			case float32:
				sum += float64(v)
			}
		}
	}
	
	return []Record{{"sum": sum}}, nil
}

// avg calcula promedio de un campo
func (a *AggregateOperator) avg(input []Record) ([]Record, error) {
	field, ok := a.params["field"].(string)
	if !ok {
		return nil, fmt.Errorf("avg requires 'field' parameter")
	}
	
	var sum float64
	count := 0
	for _, record := range input {
		if val, exists := record[field]; exists {
			switch v := val.(type) {
			case int:
				sum += float64(v)
				count++
			case int64:
				sum += float64(v)
				count++
			case float64:
				sum += v
				count++
			case float32:
				sum += float64(v)
				count++
			}
		}
	}
	
	if count == 0 {
		return []Record{{"avg": 0}}, nil
	}
	
	return []Record{{"avg": sum / float64(count)}}, nil
}

// max encuentra el valor máximo de un campo
func (a *AggregateOperator) max(input []Record) ([]Record, error) {
	field, ok := a.params["field"].(string)
	if !ok {
		return nil, fmt.Errorf("max requires 'field' parameter")
	}
	
	var max float64
	first := true
	for _, record := range input {
		if val, exists := record[field]; exists {
			var current float64
			switch v := val.(type) {
			case int:
				current = float64(v)
			case int64:
				current = float64(v)
			case float64:
				current = v
			case float32:
				current = float64(v)
			default:
				continue
			}
			
			if first || current > max {
				max = current
				first = false
			}
		}
	}
	
	return []Record{{"max": max}}, nil
}

// min encuentra el valor mínimo de un campo
func (a *AggregateOperator) min(input []Record) ([]Record, error) {
	field, ok := a.params["field"].(string)
	if !ok {
		return nil, fmt.Errorf("min requires 'field' parameter")
	}
	
	var min float64
	first := true
	for _, record := range input {
		if val, exists := record[field]; exists {
			var current float64
			switch v := val.(type) {
			case int:
				current = float64(v)
			case int64:
				current = float64(v)
			case float64:
				current = v
			case float32:
				current = float64(v)
			default:
				continue
			}
			
			if first || current < min {
				min = current
				first = false
			}
		}
	}
	
	return []Record{{"min": min}}, nil
}

// groupBy agrupa registros por clave
func (a *AggregateOperator) groupBy(input []Record) ([]Record, error) {
	keyField, ok := a.params["key"].(string)
	if !ok {
		return nil, fmt.Errorf("group_by requires 'key' parameter")
	}
	
	groups := make(map[interface{}][]Record)
	for _, record := range input {
		if key, exists := record[keyField]; exists {
			groups[key] = append(groups[key], record)
		}
	}
	
	results := make([]Record, 0, len(groups))
	for key, records := range groups {
		results = append(results, Record{
			keyField: key,
			"count":  len(records),
			"records": records,
		})
	}
	
	return results, nil
}