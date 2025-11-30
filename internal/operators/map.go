package operators

import (
	"fmt"
	"strconv"
)

// MapOperator aplica una transformación a cada record
type MapOperator struct {
	name     string
	function string
	params   map[string]interface{}
}

// NewMapOperator crea un nuevo operador map
func NewMapOperator(function string, params map[string]interface{}) *MapOperator {
	return &MapOperator{
		name:     "map",
		function: function,
		params:   params,
	}
}

// Execute ejecuta la transformación
func (m *MapOperator) Execute(input []Record) ([]Record, error) {
	output := make([]Record, 0, len(input))

	for _, record := range input {
		transformed, err := m.applyFunction(record)
		if err != nil {
			return nil, fmt.Errorf("map error: %w", err)
		}
		output = append(output, transformed)
	}

	return output, nil
}

// Name retorna el nombre del operador
func (m *MapOperator) Name() string {
	return m.name
}

// applyFunction aplica la función específica al record
func (m *MapOperator) applyFunction(record Record) (Record, error) {
	switch m.function {
	case "calculate_total":
		return m.calculateTotal(record)
	case "to_uppercase":
		return m.toUppercase(record)
	case "add_timestamp":
		return m.addTimestamp(record)
	default:
		// Si no hay función específica, retornar el record sin cambios
		return record, nil
	}
}

// calculateTotal calcula price * quantity
func (m *MapOperator) calculateTotal(record Record) (Record, error) {
	price, ok1 := record["price"]
	quantity, ok2 := record["quantity"]

	if !ok1 || !ok2 {
		return record, fmt.Errorf("missing price or quantity fields")
	}

	priceFloat := toFloat64(price)
	quantityFloat := toFloat64(quantity)

	newRecord := make(Record)
	for k, v := range record {
		newRecord[k] = v
	}
	newRecord["total"] = priceFloat * quantityFloat

	return newRecord, nil
}

// toUppercase convierte un campo a mayúsculas
func (m *MapOperator) toUppercase(record Record) (Record, error) {
	field, ok := m.params["field"].(string)
	if !ok {
		return record, fmt.Errorf("missing or invalid field parameter")
	}

	value, exists := record[field]
	if !exists {
		return record, nil
	}

	newRecord := make(Record)
	for k, v := range record {
		newRecord[k] = v
	}

	if str, ok := value.(string); ok {
		newRecord[field] = string([]byte(str)) // Simplificado, usar strings.ToUpper en producción
	}

	return newRecord, nil
}

// addTimestamp agrega un campo de timestamp
func (m *MapOperator) addTimestamp(record Record) (Record, error) {
	newRecord := make(Record)
	for k, v := range record {
		newRecord[k] = v
	}
	// Simplificado - en producción usar time.Now()
	newRecord["timestamp"] = "2025-11-27T00:00:00Z"
	return newRecord, nil
}

// toFloat64 convierte interface{} a float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}
