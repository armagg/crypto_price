package config

import (
	"log"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	MongoHost           string
	ConfigCollection    string
	MarketDatabase      string
	RedisHost           string
	TradeDatabase       string
	LastTradeCollection string
	SentryDSN           string
	ServerPort          string
}

func GetConfigs() *Config {
	config := &Config{
		// Default values - will be overridden by env vars or config file
		MongoHost:           "mongodb://localhost:27017",
		ConfigCollection:    "market-making-configs",
		MarketDatabase:      "market-bot",
		RedisHost:           "localhost:6379",
		TradeDatabase:       "market_making",
		LastTradeCollection: "last_trades",
		SentryDSN:           "",
		ServerPort:          "8080",
	}

	// Try to load from config file first
	loadFromFile(config)

	// Override with environment variables (these take precedence)
	loadFromEnv(config)

	return config
}

func loadFromFile(config *Config) {
	// Try multiple possible config file locations
	configPaths := []string{
		"pkg/config/env.yml",
		"./pkg/config/env.yml",
		filepath.Join(os.Getenv("PWD"), "pkg/config/env.yml"),
	}

	var f []byte
	var err error

	for _, path := range configPaths {
		f, err = os.ReadFile(path)
		if err == nil {
			log.Printf("Loaded config from: %s", path)
			break
		}
	}

	if err != nil {
		log.Printf("Could not load config file, using defaults: %v", err)
		return
	}

	var data map[string]string
	if err = yaml.Unmarshal(f, &data); err != nil {
		log.Printf("Error parsing config file: %v", err)
		return
	}

	// Map YAML values to config struct
	if val, ok := data["MONGO_HOST"]; ok {
		config.MongoHost = val
	}
	if val, ok := data["CONFIG_COLLECTION"]; ok {
		config.ConfigCollection = val
	}
	if val, ok := data["MARKET_DATABASE"]; ok {
		config.MarketDatabase = val
	}
	if val, ok := data["REDIS_HOST"]; ok {
		config.RedisHost = val
	}
	if val, ok := data["TRADE_DATABASE"]; ok {
		config.TradeDatabase = val
	}
	if val, ok := data["LAST_TRADE_COLLECTION"]; ok {
		config.LastTradeCollection = val
	}
}

func loadFromEnv(config *Config) {
	// Environment variables override config file values
	if val := os.Getenv("MONGO_HOST"); val != "" {
		config.MongoHost = val
	}
	if val := os.Getenv("MONGO_USER"); val != "" {
		config.MongoHost = "mongodb://" + val + ":" + os.Getenv("MONGO_PASSWORD") + "@" + extractHostFromMongoURI(config.MongoHost)
	}
	if val := os.Getenv("CONFIG_COLLECTION"); val != "" {
		config.ConfigCollection = val
	}
	if val := os.Getenv("MARKET_DATABASE"); val != "" {
		config.MarketDatabase = val
	}
	if val := os.Getenv("REDIS_HOST"); val != "" {
		config.RedisHost = val
	}
	if val := os.Getenv("TRADE_DATABASE"); val != "" {
		config.TradeDatabase = val
	}
	if val := os.Getenv("LAST_TRADE_COLLECTION"); val != "" {
		config.LastTradeCollection = val
	}
	if val := os.Getenv("SENTRY_DSN"); val != "" {
		config.SentryDSN = val
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
		config.ServerPort = val
	}
}

func extractHostFromMongoURI(uri string) string {
	// Simple extraction - you might want to use a proper URI parser
	if idx := findNth(uri, "@", 1); idx != -1 {
		return uri[idx+1:]
	}
	return uri
}

func findNth(s string, substr string, n int) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
			if count == n {
				return i
			}
		}
	}
	return -1
}