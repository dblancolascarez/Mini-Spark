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
	if len(input) == 0 {
		return []Record{}, nil
	}
	
	switch r.function {
	case "count":
		return r.count(input)
	case "sum":
		return r.sum(input)
	case "avg", "average":
		return r.avg(input)
	case "max":
		return r.max(input)
	case "min":
		return r.min(input)
	case "collect":
		return r.collect(input)
	default:
		return nil, fmt.Errorf("reduce function %s not yet implemented", r.function)
	}
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

// sum suma valores por clave
func (r *ReduceOperator) sum(input []Record) ([]Record, error) {
	valueField, ok := r.params["value_field"].(string)
	if !ok {
		return nil, fmt.Errorf("sum requires 'value_field' parameter")
	}
	
	sums := make(map[interface{}]float64)
	
	for _, record := range input {
		key, keyExists := record[r.keyField]
		if !keyExists {
			continue
		}
		
		val, valExists := record[valueField]
		if !valExists {
			continue
		}
		
		switch v := val.(type) {
		case int:
			sums[key] += float64(v)
		case int64:
			sums[key] += float64(v)
		case float64:
			sums[key] += v
		case float32:
			sums[key] += float64(v)
		}
	}
	
	results := make([]Record, 0, len(sums))
	for key, sum := range sums {
		results = append(results, Record{
			r.keyField: key,
			"sum":      sum,
		})
	}
	
	return results, nil
}

// avg calcula promedio por clave
func (r *ReduceOperator) avg(input []Record) ([]Record, error) {
	valueField, ok := r.params["value_field"].(string)
	if !ok {
		return nil, fmt.Errorf("avg requires 'value_field' parameter")
	}
	
	sums := make(map[interface{}]float64)
	counts := make(map[interface{}]int)
	
	for _, record := range input {
		key, keyExists := record[r.keyField]
		if !keyExists {
			continue
		}
		
		val, valExists := record[valueField]
		if !valExists {
			continue
		}
		
		switch v := val.(type) {
		case int:
			sums[key] += float64(v)
			counts[key]++
		case int64:
			sums[key] += float64(v)
			counts[key]++
		case float64:
			sums[key] += v
			counts[key]++
		case float32:
			sums[key] += float64(v)
			counts[key]++
		}
	}
	
	results := make([]Record, 0, len(sums))
	for key, sum := range sums {
		count := counts[key]
		avg := sum / float64(count)
		results = append(results, Record{
			r.keyField: key,
			"avg":      avg,
			"count":    count,
		})
	}
	
	return results, nil
}

// max encuentra valor máximo por clave
func (r *ReduceOperator) max(input []Record) ([]Record, error) {
	valueField, ok := r.params["value_field"].(string)
	if !ok {
		return nil, fmt.Errorf("max requires 'value_field' parameter")
	}
	
	maxValues := make(map[interface{}]float64)
	firstSeen := make(map[interface{}]bool)
	
	for _, record := range input {
		key, keyExists := record[r.keyField]
		if !keyExists {
			continue
		}
		
		val, valExists := record[valueField]
		if !valExists {
			continue
		}
		
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
		
		if !firstSeen[key] || current > maxValues[key] {
			maxValues[key] = current
			firstSeen[key] = true
		}
	}
	
	results := make([]Record, 0, len(maxValues))
	for key, max := range maxValues {
		results = append(results, Record{
			r.keyField: key,
			"max":      max,
		})
	}
	
	return results, nil
}

// min encuentra valor mínimo por clave
func (r *ReduceOperator) min(input []Record) ([]Record, error) {
	valueField, ok := r.params["value_field"].(string)
	if !ok {
		return nil, fmt.Errorf("min requires 'value_field' parameter")
	}
	
	minValues := make(map[interface{}]float64)
	firstSeen := make(map[interface{}]bool)
	
	for _, record := range input {
		key, keyExists := record[r.keyField]
		if !keyExists {
			continue
		}
		
		val, valExists := record[valueField]
		if !valExists {
			continue
		}
		
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
		
		if !firstSeen[key] || current < minValues[key] {
			minValues[key] = current
			firstSeen[key] = true
		}
	}
	
	results := make([]Record, 0, len(minValues))
	for key, min := range minValues {
		results = append(results, Record{
			r.keyField: key,
			"min":      min,
		})
	}
	
	return results, nil
}

// collect agrupa todos los valores por clave
func (r *ReduceOperator) collect(input []Record) ([]Record, error) {
	valueField, ok := r.params["value_field"].(string)
	if !ok {
		// Si no hay value_field, colectar todos los registros
		groups := make(map[interface{}][]Record)
		for _, record := range input {
			key, exists := record[r.keyField]
			if !exists {
				continue
			}
			groups[key] = append(groups[key], record)
		}
		
		results := make([]Record, 0, len(groups))
		for key, records := range groups {
			results = append(results, Record{
				r.keyField: key,
				"values":   records,
				"count":    len(records),
			})
		}
		return results, nil
	}
	
	// Colectar valores específicos
	groups := make(map[interface{}][]interface{})
	for _, record := range input {
		key, keyExists := record[r.keyField]
		if !keyExists {
			continue
		}
		
		val, valExists := record[valueField]
		if !valExists {
			continue
		}
		
		groups[key] = append(groups[key], val)
	}
	
	results := make([]Record, 0, len(groups))
	for key, values := range groups {
		results = append(results, Record{
			r.keyField: key,
			"values":   values,
			"count":    len(values),
		})
	}
	
	return results, nil
}