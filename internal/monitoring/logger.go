package monitoring

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel representa el nivel de logging
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String convierte LogLevel a string
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry representa una entrada de log estructurada
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      string                 `json:"level"`
	Component  string                 `json:"component"`
	Message    string                 `json:"message"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

// Logger maneja el logging estructurado del sistema
type Logger struct {
	component string
	level     LogLevel
	output    io.Writer
	format    string // "json" o "text"
	mu        sync.Mutex
}

// NewLogger crea un nuevo logger
func NewLogger(component string, level LogLevel, format string) *Logger {
	return &Logger{
		component: component,
		level:     level,
		output:    os.Stdout,
		format:    format,
	}
}

// SetOutput establece el destino de los logs
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// SetLevel establece el nivel mínimo de logging
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Debug registra un mensaje de debug
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.log(DEBUG, message, nil, fields...)
}

// Info registra un mensaje informativo
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.log(INFO, message, nil, fields...)
}

// Warn registra una advertencia
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.log(WARN, message, nil, fields...)
}

// Error registra un error
func (l *Logger) Error(message string, err error, fields ...map[string]interface{}) {
	l.log(ERROR, message, err, fields...)
}

// Fatal registra un error fatal y termina el programa
func (l *Logger) Fatal(message string, err error, fields ...map[string]interface{}) {
	l.log(FATAL, message, err, fields...)
	os.Exit(1)
}

// log es el método interno que escribe los logs
func (l *Logger) log(level LogLevel, message string, err error, fields ...map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Filtrar por nivel
	if level < l.level {
		return
	}

	// Crear entrada de log
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level.String(),
		Component: l.component,
		Message:   message,
	}

	// Agregar campos adicionales
	if len(fields) > 0 {
		entry.Fields = fields[0]
	}

	// Agregar error si existe
	if err != nil {
		entry.Error = err.Error()
	}

	// Escribir según formato
	if l.format == "json" {
		l.writeJSON(entry)
	} else {
		l.writeText(entry)
	}
}

// writeJSON escribe el log en formato JSON
func (l *Logger) writeJSON(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}
	fmt.Fprintln(l.output, string(data))
}

// writeText escribe el log en formato texto legible
func (l *Logger) writeText(entry LogEntry) {
	output := fmt.Sprintf("[%s] %s [%s] %s",
		entry.Timestamp,
		entry.Level,
		entry.Component,
		entry.Message,
	)

	if len(entry.Fields) > 0 {
		fieldsJSON, _ := json.Marshal(entry.Fields)
		output += fmt.Sprintf(" | fields=%s", string(fieldsJSON))
	}

	if entry.Error != "" {
		output += fmt.Sprintf(" | error=%s", entry.Error)
	}

	fmt.Fprintln(l.output, output)
}

// WithFields crea un logger temporal con campos adicionales
func (l *Logger) WithFields(fields map[string]interface{}) *LoggerContext {
	return &LoggerContext{
		logger: l,
		fields: fields,
	}
}

// LoggerContext es un logger con contexto adicional
type LoggerContext struct {
	logger *Logger
	fields map[string]interface{}
}

// Debug registra con contexto
func (lc *LoggerContext) Debug(message string) {
	lc.logger.Debug(message, lc.fields)
}

// Info registra con contexto
func (lc *LoggerContext) Info(message string) {
	lc.logger.Info(message, lc.fields)
}

// Warn registra con contexto
func (lc *LoggerContext) Warn(message string) {
	lc.logger.Warn(message, lc.fields)
}

// Error registra con contexto
func (lc *LoggerContext) Error(message string, err error) {
	lc.logger.Error(message, err, lc.fields)
}

// GlobalLogger es el logger global del sistema
var GlobalLogger *Logger

func init() {
	// Inicializar logger global con configuración por defecto
	GlobalLogger = NewLogger("mini-spark", INFO, "text")
}