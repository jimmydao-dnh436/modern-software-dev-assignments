package auth

import (
	"net/http"
	"strings"
)

type GitHubVerifier struct {
	Enabled bool
}

func NewGitHubVerifier(enabled bool) *GitHubVerifier {
	return &GitHubVerifier{Enabled: enabled}
}

func (v *GitHubVerifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !v.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "Invalid Authorization format"}`, http.StatusUnauthorized)
			return
		}
		token := parts[1]

		// Call GitHub API to verify the token
		req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
		if err != nil {
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, `{"error": "Invalid or expired GitHub token"}`, http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		next.ServeHTTP(w, r)
	})
}
