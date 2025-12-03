package operators

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
	// Para join necesitamos dos datasets
	// Por ahora asumimos que input contiene registros de ambos con un campo "_source"
	// En un sistema real, esto vendría de dos operadores diferentes
	
	leftRecords := make([]Record, 0)
	rightRecords := make([]Record, 0)
	
	// Separar registros por fuente (si existe el campo _source)
	for _, record := range input {
		if source, ok := record["_source"]; ok {
			if source == "left" {
				leftRecords = append(leftRecords, record)
			} else if source == "right" {
				rightRecords = append(rightRecords, record)
			}
		} else {
			// Si no hay _source, asumimos que son del left
			leftRecords = append(leftRecords, record)
		}
	}
	
	// Indexar right por clave para join eficiente
	rightIndex := make(map[interface{}][]Record)
	for _, record := range rightRecords {
		if key, exists := record[j.rightKey]; exists {
			rightIndex[key] = append(rightIndex[key], record)
		}
	}
	
	results := make([]Record, 0)
	
	// Realizar join
	for _, leftRecord := range leftRecords {
		leftKeyValue, exists := leftRecord[j.leftKey]
		if !exists {
			if j.joinType == "left" {
				results = append(results, leftRecord)
			}
			continue
		}
		
		rightMatches, found := rightIndex[leftKeyValue]
		
		if found {
			// Inner/Left join con matches
			for _, rightRecord := range rightMatches {
				joined := make(Record)
				// Copiar campos del left
				for k, v := range leftRecord {
					if k != "_source" {
						joined["left_"+k] = v
					}
				}
				// Copiar campos del right
				for k, v := range rightRecord {
					if k != "_source" {
						joined["right_"+k] = v
					}
				}
				results = append(results, joined)
			}
		} else if j.joinType == "left" {
			// Left join sin matches
			joined := make(Record)
			for k, v := range leftRecord {
				if k != "_source" {
					joined["left_"+k] = v
				}
			}
			results = append(results, joined)
		}
	}
	
	// Para right join, agregar registros no matcheados del right
	if j.joinType == "right" {
		// Identificar claves ya procesadas
		processedKeys := make(map[interface{}]bool)
		for _, leftRecord := range leftRecords {
			if key, exists := leftRecord[j.leftKey]; exists {
				processedKeys[key] = true
			}
		}
		
		for key, rightRecords := range rightIndex {
			if !processedKeys[key] {
				for _, rightRecord := range rightRecords {
					joined := make(Record)
					for k, v := range rightRecord {
						if k != "_source" {
							joined["right_"+k] = v
						}
					}
					results = append(results, joined)
				}
			}
		}
	}
	
	return results, nil
}

// Name retorna el nombre del operador
func (j *JoinOperator) Name() string {
	return j.name
}