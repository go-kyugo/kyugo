package kyugo

import (
	"encoding/json"
	"os"
)

func Load(path string, v interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func LoadDefault(v interface{}) error {
	return Load("config.json", v)
}

func MustLoad(path string, v interface{}) {
	if err := Load(path, v); err != nil {
		panic(err)
	}
}

type AppConfig struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
	Language    string `json:"language"`
}

type ServerConfig struct {
	Host                string     `json:"host"`
	Port                int        `json:"port"`
	ReadTimeoutSeconds  int        `json:"read_timeout_seconds"`
	WriteTimeoutSeconds int        `json:"write_timeout_seconds"`
	MaxUploadSizeBytes  int64      `json:"max_upload_size_bytes"`
	Cors                CorsConfig `json:"cors,omitempty"`
}

type CorsConfig struct {
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	AllowedMethods []string `json:"allowed_methods,omitempty"`
	AllowedHeaders []string `json:"allowed_headers,omitempty"`
}

type DatabaseConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

type Config struct {
	App      AppConfig      `json:"app"`
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
}

var ConfigVar Config

func LoadConfig(path string) error {
	return Load(path, &ConfigVar)
}

func LoadDefaultConfig() error {
	return LoadConfig("config.json")
}

func MustLoadConfig(path string) {
	if err := LoadConfig(path); err != nil {
		panic(err)
	}
}
