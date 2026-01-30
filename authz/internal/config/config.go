package config

import "os"

type Config struct {
	Port           string
	JWTSecret      []byte
	APIKeySecret   []byte
	DatabaseURL    string
	OpenFGAURL     string
	OpenFGAStoreID string
	DevMode        bool
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8002"),
		JWTSecret:      []byte(getEnv("JWT_SECRET", "")),
		APIKeySecret:   []byte(getEnv("API_KEY_SECRET", "")),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		OpenFGAURL:     getEnv("OPENFGA_URL", "http://openfga:8080"),
		OpenFGAStoreID: getEnv("OPENFGA_STORE_ID", ""),
		DevMode:        getEnv("DEV_MODE", "false") == "true",
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
