package config

import (
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// структура для конфигурации
type Config struct {
	Port     int      `yaml:"port"`
	Backends []string `yaml:"backends"`
}

// функция для загрузки конфигурации
func LoadConfig(path string) (*Config, error) {
	//открываем файл
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//читаем содержимое
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// декодируем yaml файл в структуру
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// если нет серверов в конфигурации, то возвращаем ошибку
	if len(cfg.Backends) == 0 {
		return nil, errors.New("должен быть хотя бы 1 сервер")
	}

	// возвращаем ошибку при указании некорректного порта в конфигурации
	if cfg.Port < 1024 || cfg.Port > 65535 {
		return nil, errors.New("некорректный порт")
	}

	return &cfg, nil
}
