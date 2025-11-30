// Operador flat_map aplica una función que devuelve múltiples registros por cada registro de entrada.
package operators

import (
	"fmt"
	"strings"
)

// FlatMapOperator aplica una función que genera múltiples records
type FlatMapOperator struct {
	name     string
	function string
	params   map[string]interface{}
}

// NewFlatMapOperator crea un nuevo operador flat_map
func NewFlatMapOperator(function string, params map[string]interface{}) *FlatMapOperator {
	return &FlatMapOperator{
		name:     "flat_map",
		function: function,
		params:   params,
	}
}

// Execute ejecuta la transformación
func (fm *FlatMapOperator) Execute(input []Record) ([]Record, error) {
	output := make([]Record, 0)

	for _, record := range input {
		expanded, err := fm.applyFunction(record)
		if err != nil {
			return nil, fmt.Errorf("flat_map error: %w", err)
		}
		output = append(output, expanded...)
	}

	return output, nil
}

// Name retorna el nombre del operador
func (fm *FlatMapOperator) Name() string {
	return fm.name
}

// applyFunction aplica la función específica al record
func (fm *FlatMapOperator) applyFunction(record Record) ([]Record, error) {
	switch fm.function {
	case "split_words":
		return fm.splitWords(record)
	case "explode_array":
		return fm.explodeArray(record)
	default:
		// Identidad: retornar el record como está
		return []Record{record}, nil
	}
}

// splitWords divide un texto en palabras individuales
func (fm *FlatMapOperator) splitWords(record Record) ([]Record, error) {
	text, ok := record["text"].(string)
	if !ok {
		return nil, fmt.Errorf("field 'text' not found or not a string")
	}

	words := strings.Fields(text)
	results := make([]Record, 0, len(words))

	for _, word := range words {
		// Limpiar puntuación básica
		word = strings.Trim(word, ".,!?;:\"'")
		if word != "" {
			results = append(results, Record{
				"word": strings.ToLower(word),
			})
		}
	}

	return results, nil
}

// explodeArray expande un array en múltiples records
func (fm *FlatMapOperator) explodeArray(record Record) ([]Record, error) {
	field, ok := fm.params["field"].(string)
	if !ok {
		return nil, fmt.Errorf("explode_array requires 'field' parameter")
	}

	arr, ok := record[field].([]interface{})
	if !ok {
		return []Record{record}, nil
	}

	results := make([]Record, 0, len(arr))
	for _, item := range arr {
		newRecord := make(Record)
		for k, v := range record {
			if k != field {
				newRecord[k] = v
			}
		}
		newRecord[field] = item
		results = append(results, newRecord)
	}

	return results, nil
}