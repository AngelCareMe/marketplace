package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Logger LoggerConfig `mapstructure:"logger"`
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	JWT    JWTConfig    `mapstructure:"jwt"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type DBConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	SSLMode  string `mapstructure:"sslmode"`
}

type JWTConfig struct {
	SecretKey string `mapstructure:"secret_key"`
	ExpiresIn int    `mapstructure:"expires_in"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
