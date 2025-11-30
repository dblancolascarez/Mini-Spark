// Escritura de resultados
package operators

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// WriterOperator escribe datos a archivos
type WriterOperator struct {
	name   string
	format string
	path   string
	params map[string]interface{}
}

// NewCSVWriter crea un escritor CSV
func NewCSVWriter(path string, params map[string]interface{}) *WriterOperator {
	return &WriterOperator{
		name:   "write_csv",
		format: "csv",
		path:   path,
		params: params,
	}
}

// NewJSONLWriter crea un escritor JSONL
func NewJSONLWriter(path string, params map[string]interface{}) *WriterOperator {
	return &WriterOperator{
		name:   "write_jsonl",
		format: "jsonl",
		path:   path,
		params: params,
	}
}

// Execute escribe los records al archivo
func (w *WriterOperator) Execute(input []Record) ([]Record, error) {
	switch w.format {
	case "csv":
		return input, w.writeCSV(input)
	case "jsonl":
		return input, w.writeJSONL(input)
	default:
		return nil, fmt.Errorf("unsupported format: %s", w.format)
	}
}

// Name retorna el nombre del operador
func (w *WriterOperator) Name() string {
	return w.name
}

// writeCSV escribe records a un archivo CSV
func (w *WriterOperator) writeCSV(records []Record) error {
	if len(records) == 0 {
		return fmt.Errorf("no records to write")
	}

	// Crear directorio si no existe
	if err := os.MkdirAll("data/output", 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(w.path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", w.path, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Obtener headers (claves del primer record, ordenadas)
	headers := make([]string, 0, len(records[0]))
	for key := range records[0] {
		headers = append(headers, key)
	}
	sort.Strings(headers)

	// Escribir header
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Escribir records
	for _, record := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = fmt.Sprintf("%v", record[header])
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	fmt.Printf("[Writer] Wrote %d records to %s\n", len(records), w.path)
	return nil
}

// writeJSONL escribe records a un archivo JSONL
func (w *WriterOperator) writeJSONL(records []Record) error {
	if len(records) == 0 {
		return fmt.Errorf("no records to write")
	}

	// Crear directorio si no existe
	if err := os.MkdirAll("data/output", 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(w.path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", w.path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return fmt.Errorf("failed to encode record: %w", err)
		}
	}

	fmt.Printf("[Writer] Wrote %d records to %s\n", len(records), w.path)
	return nil
}