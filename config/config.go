package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	HTTPPort    string
	DatabaseURL string
}

type postgresConfig struct {
	UserDB   string `json:"userDB"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	SSL      bool   `json:"ssl"`
}

func LoadConfig() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		content, err := ioutil.ReadFile("config/config.json")
		if err != nil {
			return Config{}, err
		}
		var pgConfig postgresConfig
		err = json.Unmarshal(content, &pgConfig)
		if err != nil {
			return Config{}, err
		}
		sslMode := "disable"
		if pgConfig.SSL {
			sslMode = "require"
		}
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			pgConfig.UserDB,
			pgConfig.Password,
			pgConfig.Host,
			pgConfig.Port,
			pgConfig.Database,
			sslMode,
		)
	}

	return Config{
		HTTPPort:    port,
		DatabaseURL: databaseURL,
	}, nil
}
