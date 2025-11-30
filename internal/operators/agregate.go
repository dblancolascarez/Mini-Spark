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
	// TODO: Implementar en Semana 3
	return nil, fmt.Errorf("aggregate operator not yet implemented")
}

// Name retorna el nombre del operador
func (a *AggregateOperator) Name() string {
	return a.name
}