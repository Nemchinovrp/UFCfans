package config

import "os"

// Config holds application settings. Environment variables override .env values.
type Config struct {
	APIKey    string
	Addr      string
	CacheFile string
}

// Load reads .env from the working directory and applies defaults.
func Load() (Config, error) {
	if err := loadEnv(".env"); err != nil {
		return Config{}, err
	}
	return Config{APIKey: os.Getenv("CITO_API_KEY"), Addr: env("ADDR", "127.0.0.1:4453"), CacheFile: env("CACHE_FILE", "data/calendar.json")}, nil
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
