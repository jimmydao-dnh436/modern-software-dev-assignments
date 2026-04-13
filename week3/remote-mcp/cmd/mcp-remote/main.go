package main

import (
	"log"
	"net/http"
	"os"

	"remote-mcp/internal/auth"
	"remote-mcp/internal/core"
	"remote-mcp/internal/httpapi"
	"remote-mcp/internal/mcpserver"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	settings := core.LoadSettings()
	logger := log.New(os.Stderr, "[weather-mcp-go] ", log.LstdFlags)

	verifier := auth.NewGitHubVerifier(settings.GitHubAuthRequired)

	client := core.NewOpenMeteoClient(settings)
	service := core.NewWeatherService(client)
	server := mcpserver.New(mcpserver.Deps{Weather: service})
	router := httpapi.NewRouter(server, settings, verifier, logger)

	addr := settings.Host + ":" + settings.Port
	logger.Printf("remote MCP server listening on http://%s/mcp", addr)
	logger.Printf("health check available at http://%s/health", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal(err)
	}
}
