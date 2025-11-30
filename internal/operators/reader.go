// Lectura de CSV/JSONL
package operators

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

// ReaderOperator lee datos de archivos
type ReaderOperator struct {
	name   string
	format string
	path   string
	params map[string]interface{}
}

// NewCSVReader crea un lector CSV
func NewCSVReader(path string, params map[string]interface{}) *ReaderOperator {
	return &ReaderOperator{
		name:   "read_csv",
		format: "csv",
		path:   path,
		params: params,
	}
}

// NewJSONLReader crea un lector JSONL
func NewJSONLReader(path string, params map[string]interface{}) *ReaderOperator {
	return &ReaderOperator{
		name:   "read_jsonl",
		format: "jsonl",
		path:   path,
		params: params,
	}
}

// Execute lee el archivo y retorna los records
func (r *ReaderOperator) Execute(input []Record) ([]Record, error) {
	switch r.format {
	case "csv":
		return r.readCSV()
	case "jsonl":
		return r.readJSONL()
	default:
		return nil, fmt.Errorf("unsupported format: %s", r.format)
	}
}

// Name retorna el nombre del operador
func (r *ReaderOperator) Name() string {
	return r.name
}

// readCSV lee un archivo CSV
func (r *ReaderOperator) readCSV() ([]Record, error) {
	file, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", r.path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Leer header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Si se especificaron columnas, usarlas en lugar del header
	if cols, ok := r.params["columns"].([]interface{}); ok {
		header = make([]string, len(cols))
		for i, col := range cols {
			header[i] = col.(string)
		}
		// Reiniciar el lector para leer desde el inicio
		file.Seek(0, 0)
		reader = csv.NewReader(file)
	}

	records := make([]Record, 0)
	
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		record := make(Record)
		for i, value := range row {
			if i < len(header) {
				// Intentar convertir a número si es posible
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					record[header[i]] = num
				} else {
					record[header[i]] = value
				}
			}
		}
		records = append(records, record)
	}

	fmt.Printf("[Reader] Read %d records from %s\n", len(records), r.path)
	return records, nil
}

// readJSONL lee un archivo JSONL (JSON Lines)
func (r *ReaderOperator) readJSONL() ([]Record, error) {
	file, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", r.path, err)
	}
	defer file.Close()

	records := make([]Record, 0)
	decoder := json.NewDecoder(file)

	for {
		var record Record
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}
		records = append(records, record)
	}

	fmt.Printf("[Reader] Read %d records from %s\n", len(records), r.path)
	return records, nil
}
