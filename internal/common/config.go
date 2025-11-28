// Configuración global
package common

import "time"

// Config contiene la configuración global del sistema
type Config struct {
	Master MasterConfig
	Worker WorkerConfig
	Common CommonConfig
}

type MasterConfig struct {
	Host              string
	Port              int
	HeartbeatTimeout  time.Duration
	SchedulerPolicy   string // "round-robin", "least-loaded"
	PersistenceType   string // "file", "sqlite"
}

type WorkerConfig struct {
	MasterAddress     string
	Port              int
	HeartbeatInterval time.Duration
	TaskPoolSize      int
	MemoryLimitMB     int
	SpillThresholdMB  int
}

type CommonConfig struct {
	LogLevel          string
	DataDir           string
	TempDir           string
}

// DefaultConfig retorna configuración por defecto
func DefaultConfig() *Config {
	return &Config{
		Master: MasterConfig{
			Host:              "localhost",
			Port:              8080,
			HeartbeatTimeout:  10 * time.Second,
			SchedulerPolicy:   "round-robin",
			PersistenceType:   "file",
		},
		Worker: WorkerConfig{
			MasterAddress:     "localhost:8080",
			Port:              8081,
			HeartbeatInterval: 3 * time.Second,
			TaskPoolSize:      4,
			MemoryLimitMB:     512,
			SpillThresholdMB:  256,
		},
		Common: CommonConfig{
			LogLevel:          "info",
			DataDir:           "./data",
			TempDir:           "/tmp/mini-spark",
		},
	}
}
