package config

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	AppConfig
	ServerConfig
	MongoDBConfig
	WebhookConfig
	RabbitMQConfig
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

type WebhookConfig struct {
	URL string
}

type RabbitMQConfig struct {
	URL string
}

func New() (*Config, error) {
	config := &Config{}

	viper.SetConfigFile("../config/.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	mongoURL := "mongodb://localhost:27017"
	rabbitHost := "localhost"
	if os.Getenv("DOCKER_ENV") == "1" {
		rabbitHost = "rabbitmq"
		mongoURL = "mongodb://mongodb:27017/message_bird"
	}

	rabbitMQURL := fmt.Sprintf("amqp://guest:guest@%s:5672/", rabbitHost)

	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Msg("Could not read the .env config file.")
		return nil, err
	}

	requiredKeys := []string{
		"PORT",
		"MONGODB_NAME",
		"WEBHOOK_SITE_URL",
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
		Host: mongoURL,
		Name: viper.GetString("MONGODB_NAME"),
	}
	config.WebhookConfig = WebhookConfig{
		URL: viper.GetString("WEBHOOK_SITE_URL"),
	}
	config.RabbitMQConfig = RabbitMQConfig{
		URL: rabbitMQURL,
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
