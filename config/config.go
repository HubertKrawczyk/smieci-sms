package config

import "os"

type Config struct {
	ServerAddress     string
	DatabaseURL       string
	CityGarbageURL    string
	SMSProviderAPIKey string
}

func LoadConfig() Config {
	return Config{
		ServerAddress:     getEnv("SERVER_ADDRESS", ":8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:1234@localhost:5432/smieci"),
		CityGarbageURL:    getEnv("CITY_GARBAGE_URL", "https://example.com/garbage-schedule"),
		SMSProviderAPIKey: getEnv("SMS_PROVIDER_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
