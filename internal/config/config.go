package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Mode               string        `yaml:"mode"`
	AccessTokenTTL     time.Duration `yaml:"access_token_ttl" env-default:"30m"`
	RefreshTokenTTL    time.Duration `yaml:"refresh_token_ttl" env-default:"24h"`
	ActivationTokenTTL time.Duration `yaml:"activation_token_ttl" env-default:"6h"`
	TokenSigningAlg    string        `yaml:"token_signing_alg" env-required:"true"`
	Server             Server        `yaml:"server" env-required:"true"`
	DB                 DB            `yaml:"db" env-required:"true"`
}

type Server struct {
	Port    string        `yaml:"port"`
	Host    string        `yaml:"host"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
}

type DB struct {
	Port        string        `yaml:"port" env:"DB_PORT" env-default:"5432"`
	Host        string        `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Name        string        `yaml:"name" env:"DB_NAME" env-required:"true"`
	User        string        `yaml:"user" env:"DB_USER" env-required:"true"`
	Password    string        `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
	LoadTimeout time.Duration `yaml:"load_timeout" env-default:"10s"`
}

func Load(configPath string) *Config {
	if configPath == "" {
		panic("config path is required")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config on path " + configPath + " does not exist")
	}
	var cfg Config
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}
