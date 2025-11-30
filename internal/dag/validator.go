package dag

import "fmt"

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(dag *DAG) error {
	if len(dag.Nodes) == 0 {
		return fmt.Errorf("DAG is empty")
	}
	
	roots := dag.GetRootNodes()
	if len(roots) == 0 {
		return fmt.Errorf("DAG has no root nodes")
	}
	
	leaves := dag.GetLeafNodes()
	if len(leaves) == 0 {
		return fmt.Errorf("DAG has no leaf nodes")
	}
	
	_, err := dag.TopologicalSort()
	if err != nil {
		return fmt.Errorf("DAG validation failed: %w", err)
	}
	
	// Por ahora, aceptar todos los operadores
	return nil
}