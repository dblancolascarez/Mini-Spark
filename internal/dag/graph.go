// Representación del DAG
package dag

import (
	"fmt"
	"sort"
)

// DAG representa un grafo acíclico dirigido
type DAG struct {
	Nodes map[string]*Node
	Edges []*Edge
}

// Node representa un nodo (operador) en el DAG
type Node struct {
	ID         string
	Operator   string
	Parameters map[string]interface{}
	Children   []*Node
	Parents    []*Node
}

// Edge representa una arista entre nodos
type Edge struct {
	From string
	To   string
}

// NewDAG crea un nuevo DAG vacío
func NewDAG() *DAG {
	return &DAG{
		Nodes: make(map[string]*Node),
		Edges: make([]*Edge, 0),
	}
}

// AddNode agrega un nodo al DAG
func (d *DAG) AddNode(id, operator string, params map[string]interface{}) error {
	if _, exists := d.Nodes[id]; exists {
		return fmt.Errorf("node %s already exists", id)
	}

	node := &Node{
		ID:         id,
		Operator:   operator,
		Parameters: params,
		Children:   make([]*Node, 0),
		Parents:    make([]*Node, 0),
	}

	d.Nodes[id] = node
	return nil
}

// AddEdge agrega una arista al DAG
func (d *DAG) AddEdge(from, to string) error {
	fromNode, fromExists := d.Nodes[from]
	toNode, toExists := d.Nodes[to]

	if !fromExists {
		return fmt.Errorf("node %s does not exist", from)
	}
	if !toExists {
		return fmt.Errorf("node %s does not exist", to)
	}

	// Verificar que no crea ciclo
	if d.wouldCreateCycle(fromNode, toNode) {
		return fmt.Errorf("adding edge %s -> %s would create a cycle", from, to)
	}

	edge := &Edge{From: from, To: to}
	d.Edges = append(d.Edges, edge)

	fromNode.Children = append(fromNode.Children, toNode)
	toNode.Parents = append(toNode.Parents, fromNode)

	return nil
}

// wouldCreateCycle verifica si agregar una arista crearía un ciclo
func (d *DAG) wouldCreateCycle(from, to *Node) bool {
	visited := make(map[string]bool)
	return d.hasCycle(to, from.ID, visited)
}

// hasCycle realiza DFS para detectar ciclos
func (d *DAG) hasCycle(node *Node, targetID string, visited map[string]bool) bool {
	if node.ID == targetID {
		return true
	}

	if visited[node.ID] {
		return false
	}

	visited[node.ID] = true

	for _, child := range node.Children {
		if d.hasCycle(child, targetID, visited) {
			return true
		}
	}

	return false
}

// TopologicalSort retorna los nodos en orden topológico
func (d *DAG) TopologicalSort() ([]*Node, error) {
	inDegree := make(map[string]int)
	for id := range d.Nodes {
		inDegree[id] = 0
	}

	for _, edge := range d.Edges {
		inDegree[edge.To]++
	}

	queue := make([]*Node, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, d.Nodes[id])
		}
	}

	result := make([]*Node, 0)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, child := range node.Children {
			inDegree[child.ID]--
			if inDegree[child.ID] == 0 {
				queue = append(queue, child)
			}
		}
	}

	if len(result) != len(d.Nodes) {
		return nil, fmt.Errorf("DAG contains a cycle")
	}

	return result, nil
}

// GetRootNodes retorna los nodos sin padres
func (d *DAG) GetRootNodes() []*Node {
	roots := make([]*Node, 0)
	for _, node := range d.Nodes {
		if len(node.Parents) == 0 {
			roots = append(roots, node)
		}
	}
	
	// Ordenar por ID para consistencia
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].ID < roots[j].ID
	})
	
	return roots
}

// GetLeafNodes retorna los nodos sin hijos
func (d *DAG) GetLeafNodes() []*Node {
	leaves := make([]*Node, 0)
	for _, node := range d.Nodes {
		if len(node.Children) == 0 {
			leaves = append(leaves, node)
		}
	}
	
	sort.Slice(leaves, func(i, j int) bool {
		return leaves[i].ID < leaves[j].ID
	})
	
	return leaves
}

// GetStages retorna los nodos agrupados por nivel (para ejecución paralela)
func (d *DAG) GetStages() ([][]*Node, error) {
	sorted, err := d.TopologicalSort()
	if err != nil {
		return nil, err
	}

	levels := make(map[string]int)
	for _, node := range sorted {
		maxParentLevel := -1
		for _, parent := range node.Parents {
			if levels[parent.ID] > maxParentLevel {
				maxParentLevel = levels[parent.ID]
			}
		}
		levels[node.ID] = maxParentLevel + 1
	}

	maxLevel := 0
	for _, level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	stages := make([][]*Node, maxLevel+1)
	for i := range stages {
		stages[i] = make([]*Node, 0)
	}

	for _, node := range sorted {
		level := levels[node.ID]
		stages[level] = append(stages[level], node)
	}

	return stages, nil
}