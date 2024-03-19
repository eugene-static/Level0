package config

import (
	"os"
	"time"

	"github.com/go-yaml/yaml"
)

type Config struct {
	Postgres Postgres `yaml:"postgres"`
	Nats     Nats     `yaml:"nats"`
	Server   Server   `yaml:"server"`
	Logger   Logger   `yaml:"logger"`
}

type Postgres struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
	SSL      string `yaml:"ssl_mode"`
}

type Nats struct {
	ClusterID string `yaml:"cluster_id"`
	ClientID  string `yaml:"client_id"`
	URL       string `yaml:"url"`
	Subject   string `yaml:"subject"`
}

type Server struct {
	Network      string        `yaml:"network"`
	IP           string        `yaml:"IP"`
	Port         string        `yaml:"port"`
	Timeout      time.Duration `yaml:"timeout"`
	Shutdown     time.Duration `yaml:"shutdown"`
	TemplatePath string        `yaml:"template_path"`
	StaticPath   string        `yaml:"static_path"`
}

type Logger struct {
	Level   string `yaml:"level"`
	Handler string `yaml:"handler"`
}

func Get(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	cfg := &Config{}
	if err = yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
