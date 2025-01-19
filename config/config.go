package config

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	AppConfig
	ServerConfig
	MongoDBConfig
}

type AppConfig struct {
	AppName string
}

type ServerConfig struct {
	Port int
}

type MongoDBConfig struct {
	Host string
	Name string
}

func New() (*Config, error) {
	config := &Config{}

	viper.SetConfigFile("../config/.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Msg("Could not read the .env config file.")
		return nil, err
	}

	requiredKeys := []string{
		"PORT",
		"MONGODB_HOST",
		"MONGODB_NAME",
	}

	missingKeys := checkMissingKeys(requiredKeys)
	if len(missingKeys) > 0 {
		errMsg := fmt.Sprintf("Missing required configuration keys: %v", missingKeys)
		log.Error().Msg(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	config.AppConfig = AppConfig{
		AppName: viper.GetString("APPNAME"),
	}
	config.ServerConfig = ServerConfig{
		Port: viper.GetInt("PORT"),
	}
	config.MongoDBConfig = MongoDBConfig{
		Host: viper.GetString("MONGODB_HOST"),
		Name: viper.GetString("MONGODB_NAME"),
	}

	return config, nil
}

func checkMissingKeys(keys []string) []string {
	var missingKeys []string
	for _, key := range keys {
		if !viper.IsSet(key) || viper.GetString(key) == "" {
			missingKeys = append(missingKeys, key)
		}
	}
	return missingKeys
}
