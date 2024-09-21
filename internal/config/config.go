package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	*ApplicationConfig
}

type ApplicationConfig struct {
	Server   *ServerConfig      `mapstructure:"server"`
	NewRelic *NewRelicConfig    `mapstructure:"newrelic"`
	Clusters map[string]Cluster `mapstructure:"clusters"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type NewRelicConfig struct {
	Name    string `mapstructure:"name"`
	License string `mapstructure:"license"`
}

type QueryParam struct {
	timeRange string
}

type Cluster struct {
	Brokers []string `mapstructure:"brokers"`
}

func NewConfig(configFile, secretFile string) (*ApplicationConfig, error) {
	filename := filepath.Base(configFile)
	ext := filepath.Ext(configFile)
	configPath := filepath.Dir(configFile)

	viper.SetConfigType(strings.TrimPrefix(ext, "."))
	viper.SetConfigName(strings.TrimSuffix(filename, ext))
	viper.AddConfigPath(configPath)

	_, _ = fmt.Fprintf(os.Stdout, "reading config '%s'\n", configFile)

	if err := viper.ReadInConfig(); err != nil {
		return &ApplicationConfig{}, err
	}

	if err := handleSecrets(secretFile); err != nil {
		return &ApplicationConfig{}, err
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var cfg *ApplicationConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return &ApplicationConfig{}, err
	}

	return cfg, nil
}

func handleSecrets(secretFile string) error {
	if _, err := os.Stat(secretFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// if file does not exist, skip it
			return nil
		}
		return err
	}

	filename := filepath.Base(secretFile)
	ext := filepath.Ext(secretFile)
	secretPath := filepath.Dir(secretFile)

	viper.SetConfigName(strings.TrimSuffix(filename, ext))
	viper.AddConfigPath(secretPath)

	_, _ = fmt.Fprintf(os.Stdout, "reading secret '%s'\n", secretFile)

	return viper.MergeInConfig()
}
