package config

import (
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
		log.Fatal().Err(err)
		return nil, err
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
