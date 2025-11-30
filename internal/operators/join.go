package operators

import (
	"fmt"
)

// JoinOperator realiza un join entre dos datasets
type JoinOperator struct {
	name     string
	joinType string // "inner", "left", "right"
	leftKey  string
	rightKey string
	params   map[string]interface{}
}

// NewJoinOperator crea un nuevo operador join
func NewJoinOperator(joinType, leftKey, rightKey string, params map[string]interface{}) *JoinOperator {
	return &JoinOperator{
		name:     "join",
		joinType: joinType,
		leftKey:  leftKey,
		rightKey: rightKey,
		params:   params,
	}
}

// Execute ejecuta el join
func (j *JoinOperator) Execute(input []Record) ([]Record, error) {
	// TODO: Implementar en Semana 3
	return nil, fmt.Errorf("join operator not yet implemented")
}

// Name retorna el nombre del operador
func (j *JoinOperator) Name() string {
	return j.name
}