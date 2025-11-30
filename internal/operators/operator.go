package operators

// Record representa un registro de datos
type Record map[string]interface{}

// Operator es la interfaz base para todos los operadores
type Operator interface {
	Execute(input []Record) ([]Record, error)
	Name() string
}

// MapFunc es una función que transforma un record
type MapFunc func(Record) (Record, error)

// FilterFunc es una función que filtra records
type FilterFunc func(Record) bool

// FlatMapFunc es una función que genera múltiples records
type FlatMapFunc func(Record) ([]Record, error)

// ReduceFunc es una función que reduce valores por clave
type ReduceFunc func(key interface{}, values []interface{}) interface{}
