package operators

import (
	"fmt"
)

// ShuffleOperator redistribuye datos entre particiones
type ShuffleOperator struct {
	name       string
	partitions int
	params     map[string]interface{}
}

// NewShuffleOperator crea un nuevo operador shuffle
func NewShuffleOperator(partitions int, params map[string]interface{}) *ShuffleOperator {
	return &ShuffleOperator{
		name:       "shuffle",
		partitions: partitions,
		params:     params,
	}
}

// Execute ejecuta el shuffle
func (s *ShuffleOperator) Execute(input []Record) ([]Record, error) {
	// TODO: Implementar en Semana 3
	return nil, fmt.Errorf("shuffle operator not yet implemented")
}

// Name retorna el nombre del operador
func (s *ShuffleOperator) Name() string {
	return s.name
}