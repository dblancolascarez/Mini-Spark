// Parser de definición de jobs
package dag

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dblancolascarez/mini-spark/internal/protocol"
)

// Parser convierte definiciones JSON a DAG
type Parser struct{}

// NewParser crea un nuevo parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFromFile lee un archivo JSON y construye el DAG
func (p *Parser) ParseFromFile(path string) (*DAG, *protocol.JobSubmitRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.ParseFromBytes(data)
}

// ParseFromBytes parsea bytes JSON y construye el DAG
func (p *Parser) ParseFromBytes(data []byte) (*DAG, *protocol.JobSubmitRequest, error) {
	var req protocol.JobSubmitRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	dag, err := p.BuildDAG(&req.DAG)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to Build DAG: %w", err)
	}

	return dag, &req, nil
}

// BuildDAG construye el DAG a partir de la definición
func (p *Parser) BuildDAG(def *protocol.DAGDefinition) (*DAG, error) {
	dag := NewDAG()

	// Agregar nodos
	for _, nodeDef := range def.Nodes {
		err := dag.AddNode(nodeDef.ID, nodeDef.Operator, nodeDef.Parameters)
		if err != nil {
			return nil, err
		}
	}

	// Agregar aristas
	for _, edgeDef := range def.Edges {
		err := dag.AddEdge(edgeDef.From, edgeDef.To)
		if err != nil {
			return nil, err
		}
	}

	return dag, nil
}
