package core

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Settings struct {
	WeatherBaseURL          string
	WeatherGeocodingBaseURL string
	Timeout                 time.Duration
	UserAgent               string
	RetryAttempts           int
	Backoff                 time.Duration
	Host                    string
	Port                    string
	AllowedOrigins          []string

	// GitHub OAuth Settings
	GitHubAuthRequired bool
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURI  string
	GitHubOAuthScopes  []string
}

func LoadSettings() Settings {
	timeoutSeconds := getEnvFloat("WEATHER_TIMEOUT_SECONDS", 10)
	backoffSeconds := getEnvFloat("WEATHER_BACKOFF_SECONDS", 0.2)

	return Settings{
		WeatherBaseURL:          getEnv("WEATHER_BASE_URL", "https://api.open-meteo.com/v1"),
		WeatherGeocodingBaseURL: getEnv("WEATHER_GEOCODING_BASE_URL", "https://geocoding-api.open-meteo.com/v1"),
		Timeout:                 time.Duration(timeoutSeconds * float64(time.Second)),
		UserAgent:               getEnv("WEATHER_USER_AGENT", "week3-mcp-weather-server-go/1.0"),
		RetryAttempts:           getEnvInt("WEATHER_RETRY_ATTEMPTS", 2),
		Backoff:                 time.Duration(backoffSeconds * float64(time.Second)),
		Host:                    getEnv("MCP_HOST", "127.0.0.1"),
		Port:                    getEnv("MCP_PORT", "8000"),
		AllowedOrigins:          getEnvCSV("MCP_ALLOWED_ORIGINS", []string{"http://localhost:6274", "http://127.0.0.1:6274"}),

		// Load from env, matching Python's setup
		GitHubAuthRequired: getEnvBool("GITHUB_AUTH_REQUIRED", true),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURI:  getEnv("GITHUB_REDIRECT_URI", "http://127.0.0.1:9999/callback"),
		GitHubOAuthScopes:  getEnvCSV("GITHUB_OAUTH_SCOPES", []string{"read:user"}),
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvFloat(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvCSV(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	var result []string
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
