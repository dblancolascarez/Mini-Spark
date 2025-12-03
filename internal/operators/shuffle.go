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
	if len(input) == 0 {
		return input, nil
	}
	
	// Obtener clave para particionar
	keyField, ok := s.params["key"].(string)
	if !ok {
		// Si no hay clave, hacer shuffle aleatorio
		return s.randomShuffle(input)
	}
	
	// Particionar por hash de la clave
	partitions := make([][]Record, s.partitions)
	for i := range partitions {
		partitions[i] = make([]Record, 0)
	}
	
	for _, record := range input {
		if keyVal, exists := record[keyField]; exists {
			partitionIdx := s.hashPartition(keyVal, s.partitions)
			partitions[partitionIdx] = append(partitions[partitionIdx], record)
		} else {
			// Si no tiene la clave, ir a partición 0
			partitions[0] = append(partitions[0], record)
		}
	}
	
	// Combinar particiones con metadata
	result := make([]Record, 0, len(input))
	for idx, partition := range partitions {
		for _, record := range partition {
			// Agregar metadata de partición
			record["_partition"] = idx
			result = append(result, record)
		}
	}
	
	return result, nil
}

// Name retorna el nombre del operador
func (s *ShuffleOperator) Name() string {
	return s.name
}

// hashPartition calcula la partición usando hash
func (s *ShuffleOperator) hashPartition(key interface{}, numPartitions int) int {
	// Hash simple basado en string representation
	keyStr := fmt.Sprintf("%v", key)
	hash := 0
	for _, char := range keyStr {
		hash = (hash*31 + int(char)) % numPartitions
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// randomShuffle mezcla registros aleatoriamente
func (s *ShuffleOperator) randomShuffle(input []Record) ([]Record, error) {
	result := make([]Record, len(input))
	copy(result, input)
	
	// Fisher-Yates shuffle
	for i := len(result) - 1; i > 0; i-- {
		j := (i * 31) % (i + 1) // Pseudo-random
		result[i], result[j] = result[j], result[i]
	}
	
	return result, nil
}