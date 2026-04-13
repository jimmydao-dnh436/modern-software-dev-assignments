package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"remote-mcp/internal/auth"
	"remote-mcp/internal/core"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewRouter(server *mcp.Server, settings core.Settings, verifier *auth.GitHubVerifier, logger *log.Logger) http.Handler {
	mcpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, nil)

	mux := http.NewServeMux()

	var protectedMCPHandler http.Handler
	if settings.GitHubAuthRequired {
		logger.Println("auth: GitHub OAuth middleware ENABLED")
		protectedMCPHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			verifier.Middleware(mcpHandler).ServeHTTP(w, r)
		})
	} else {
		logger.Println("auth: GitHub OAuth middleware DISABLED")
		protectedMCPHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			mcpHandler.ServeHTTP(w, r)
		})
	}

	mux.Handle("/mcp", protectedMCPHandler)
	mux.Handle("/mcp/", protectedMCPHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":        "ok",
			"auth_required": settings.GitHubAuthRequired,
		})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"service": "Weather MCP (Go)"})
	})

	return requestLogger(logger, corsAndOriginGuard(settings, mux))
}

func corsAndOriginGuard(settings core.Settings, next http.Handler) http.Handler {
	allowedOrigins := make(map[string]struct{})
	for _, origin := range settings.AllowedOrigins {
		allowedOrigins[strings.TrimSpace(origin)] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if _, ok := allowedOrigins[origin]; ok || origin == "http://localhost:6274" || origin == "http://127.0.0.1:6274" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Mcp-Session-Id, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func requestLogger(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
