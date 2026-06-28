package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr               string
	ReadTimeout        int
	WriteTimeout       int
	IdleTimeout        int
	AwsAccessKey       string
	AwsSecretKey       string
	AwsRegion          string
	SlackSigningSecret string
	SlackBotToken      string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	var errs []string

	req := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			errs = append(errs, "missing required environment variable: "+key)
		}
		return v
	}

	cfg := Config{
		Addr:               optStr("ADDR", ":8080"),
		ReadTimeout:        optInt("READ_TIMEOUT", 5),
		WriteTimeout:       optInt("WRITE_TIMEOUT", 10),
		IdleTimeout:        optInt("IDLE_TIMEOUT", 120),
		AwsAccessKey:       optStr("AWS_ACCESS_KEY_ID", ""),
		AwsSecretKey:       optStr("AWS_SECRET_ACCESS_KEY", ""),
		AwsRegion:          optStr("AWS_REGION", ""),
		SlackSigningSecret: req("SLACK_SIGNING_SECRET"),
		SlackBotToken:      req("SLACK_BOT_TOKEN"),
	}

	if len(errs) > 0 {
		return Config{}, fmt.Errorf("configuration errors: %s", strings.Join(errs, "; "))
	}

	return cfg, nil
}

func optStr(key string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func optInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		_, err := fmt.Sscanf(v, "%d", &i)
		if err == nil {
			return i
		}
	}
	return def
}
