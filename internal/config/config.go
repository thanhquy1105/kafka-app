package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ProducerConfig holds producer configuration
type ProducerConfig struct {
	Brokers      []string `yaml:"brokers"`
	Topic        string   `yaml:"topic"`
	BatchSize    int      `yaml:"batch_size"`
	BatchTimeout int      `yaml:"batch_timeout_ms"`
	MaxRetries   int      `yaml:"max_retries"`
	LogLevel     string   `yaml:"log_level"`
	MetricsPort  int      `yaml:"metrics_port"`
}

// ConsumerConfig holds consumer configuration
type ConsumerConfig struct {
	Brokers        []string `yaml:"brokers"`
	Topics         []string `yaml:"topics"`
	GroupID        string   `yaml:"group_id"`
	AutoCommit     bool     `yaml:"auto_commit"`
	CommitInterval int      `yaml:"commit_interval_ms"`
	LogLevel       string   `yaml:"log_level"`
	MetricsPort    int      `yaml:"metrics_port"`
}

// LoadProducerConfig loads producer config from YAML file
func LoadProducerConfig(path string) (*ProducerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ProducerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConsumerConfig loads consumer config from YAML file
func LoadConsumerConfig(path string) (*ConsumerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ConsumerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
