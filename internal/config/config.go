package config

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Mode               string        `yaml:"mode"`
		AccessTokenTTL     time.Duration `yaml:"access_token_ttl" env-default:"30m"`
		RefreshTokenTTL    time.Duration `yaml:"refresh_token_ttl" env-default:"24h"`
		ActivationTokenTTL time.Duration `yaml:"activation_token_ttl" env-default:"30m"`
		TokenSigningAlg    string        `yaml:"token_signing_alg" env-default:"HS256"`
		Server             Server        `yaml:"server" env-required:"true"`
		DB                 DB            `yaml:"db" env-required:"true"`
	}
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	}
	DB struct {
		Dsn string `yaml:"dsn" env:"DB_DSN"`
	}
)

func MustLoad(configPath string) *Config {
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

func ResolveConfigPath() string {
	mode := os.Getenv("MODE")
	if mode == "" {
		panic("Unable to determine config path. MODE env variable is not set")
	}
	currDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to build config path. Error: %s", err))
	}
	newPath := path.Join(currDir, "config", mode)
	newPath += ".yaml"
	if _, err := os.Stat(newPath); err != nil {
		panic(fmt.Sprintf("Failed to build config path. Error: %s", err))
	}
	return newPath
}
