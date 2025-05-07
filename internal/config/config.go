package config

import (
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int      `yaml:"port"`
	Backends []string `yaml:"backends"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if len(cfg.Backends) == 0 {
		return nil, errors.New("должен быть хотя бы 1 сервер")
	}

	if cfg.Port < 1 {
		return nil, errors.New("некорректный порт")
	}

	return &cfg, nil
}
