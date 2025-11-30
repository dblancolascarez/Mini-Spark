package operators

import (
	"fmt"
	"strconv"
	"strings"
)

// FilterOperator filtra records basado en una condición
type FilterOperator struct {
	name      string
	condition string
	params    map[string]interface{}
}

// NewFilterOperator crea un nuevo operador filter
func NewFilterOperator(condition string, params map[string]interface{}) *FilterOperator {
	return &FilterOperator{
		name:      "filter",
		condition: condition,
		params:    params,
	}
}

// Execute filtra los records
func (f *FilterOperator) Execute(input []Record) ([]Record, error) {
	output := make([]Record, 0)

	for _, record := range input {
		passes, err := f.evaluateCondition(record)
		if err != nil {
			return nil, fmt.Errorf("filter error: %w", err)
		}
		if passes {
			output = append(output, record)
		}
	}

	return output, nil
}

// Name retorna el nombre del operador
func (f *FilterOperator) Name() string {
	return f.name
}

// evaluateCondition evalúa la condición para un record
func (f *FilterOperator) evaluateCondition(record Record) (bool, error) {
	// Parser simple de condiciones (e.g., "price > 100", "status == active")
	parts := strings.Fields(f.condition)
	if len(parts) < 3 {
		return false, fmt.Errorf("invalid condition format: %s", f.condition)
	}

	field := parts[0]
	operator := parts[1]
	valueStr := parts[2]

	recordValue, exists := record[field]
	if !exists {
		return false, nil
	}

	return f.compareValues(recordValue, operator, valueStr)
}

// compareValues compara dos valores según el operador
func (f *FilterOperator) compareValues(recordValue interface{}, operator, targetStr string) (bool, error) {
	switch operator {
	case ">":
		return f.compareNumeric(recordValue, targetStr, func(a, b float64) bool { return a > b })
	case ">=":
		return f.compareNumeric(recordValue, targetStr, func(a, b float64) bool { return a >= b })
	case "<":
		return f.compareNumeric(recordValue, targetStr, func(a, b float64) bool { return a < b })
	case "<=":
		return f.compareNumeric(recordValue, targetStr, func(a, b float64) bool { return a <= b })
	case "==":
		return f.compareEqual(recordValue, targetStr)
	case "!=":
		eq, err := f.compareEqual(recordValue, targetStr)
		return !eq, err
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// compareNumeric compara valores numéricos
func (f *FilterOperator) compareNumeric(recordValue interface{}, targetStr string, cmp func(float64, float64) bool) (bool, error) {
	recordFloat := toFloat64(recordValue)
	targetFloat, err := strconv.ParseFloat(targetStr, 64)
	if err != nil {
		return false, fmt.Errorf("invalid numeric value: %s", targetStr)
	}

	return cmp(recordFloat, targetFloat), nil
}

// compareEqual compara valores para igualdad
func (f *FilterOperator) compareEqual(recordValue interface{}, targetStr string) (bool, error) {
	// Remover comillas si existen
	targetStr = strings.Trim(targetStr, "\"'")

	switch v := recordValue.(type) {
	case string:
		return v == targetStr, nil
	case int:
		target, err := strconv.Atoi(targetStr)
		if err != nil {
			return false, nil
		}
		return v == target, nil
	case float64:
		target, err := strconv.ParseFloat(targetStr, 64)
		if err != nil {
			return false, nil
		}
		return v == target, nil
	default:
		return fmt.Sprintf("%v", v) == targetStr, nil
	}
}
