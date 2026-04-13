package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"

	"github.com/joho/godotenv"
)

// openBrowser giúp tự động mở link trên trình duyệt mặc định tuỳ theo Hệ điều hành
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin": // macOS
		cmd = "open"
	default: // Linux
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func main() {
	// Load file .env từ thư mục gốc
	_ = godotenv.Load()

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")
	scopes := os.Getenv("GITHUB_OAUTH_SCOPES")

	if clientID == "" || clientSecret == "" {
		log.Fatal("❌ GITHUB_CLIENT_ID và GITHUB_CLIENT_SECRET chưa được set trong file .env")
	}
	if redirectURI == "" {
		redirectURI = "http://127.0.0.1:9999/callback"
	}
	if scopes == "" {
		scopes = "read:user"
	}

	state := "mcp_go_random_state" // State giả lập bảo mật

	// Tạo Auth URL
	authURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		clientID, url.QueryEscape(redirectURI), url.QueryEscape(scopes), state)

	fmt.Println("Đang mở trình duyệt để xác thực với GitHub...")
	fmt.Println("  URL:", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Println("Không thể tự mở trình duyệt. Bạn hãy copy link trên dán vào trình duyệt nhé!")
	}

	// Lấy port từ cấu hình REDIRECT_URI để dựng server nghe ngóng (ví dụ: 9999)
	u, err := url.Parse(redirectURI)
	if err != nil {
		log.Fatal(err)
	}
	port := u.Port()
	if port == "" {
		port = "9999" // Mặc định
	}

	// Khởi tạo HTTP server cục bộ
	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":" + port, Handler: mux}
	shutdownChan := make(chan struct{})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Không nhận được code từ GitHub", http.StatusBadRequest)
			return
		}

		// Nhận được code, tiến hành đổi lấy Access Token
		tokenReq, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
		q := tokenReq.URL.Query()
		q.Add("client_id", clientID)
		q.Add("client_secret", clientSecret)
		q.Add("code", code)
		q.Add("redirect_uri", redirectURI)
		tokenReq.URL.RawQuery = q.Encode()
		tokenReq.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(tokenReq)
		if err != nil {
			http.Error(w, "Lỗi khi gọi lên GitHub lấy token", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var data map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&data)

		// Kiểm tra phản hồi từ GitHub
		if errMsg, ok := data["error"]; ok {
			fmt.Printf("\n❌ Lỗi đổi token: %v\n", errMsg)
			fmt.Fprintf(w, "Lỗi: %v", errMsg)
		} else {
			token := data["access_token"].(string)
			fmt.Println("\n============================================================")
			fmt.Println("✅ THÀNH CÔNG! Đây là GitHub Access Token của bạn:")
			fmt.Println(token)
			fmt.Println("============================================================")
			fmt.Println("\nHãy dán token này vào MCP Inspector Custom Headers như sau:")
			fmt.Println("Authorization: Bearer", token)
			fmt.Println("============================================================")

			// Trả về giao diện web thông báo thành công cho người dùng
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
				<div style="font-family: sans-serif; text-align: center; margin-top: 50px;">
					<h1 style="color: #4CAF50;">✅ Xác thực thành công!</h1>
					<p>Đã lấy được Access Token. Bạn có thể đóng tab này và quay lại terminal.</p>
				</div>
			`)
		}

		// Lấy xong token thì tự động tắt local server
		go func() {
			srv.Shutdown(context.Background())
			close(shutdownChan)
		}()
	})

	fmt.Printf("Đang chờ GitHub phản hồi (callback) tại cổng %s...\n", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Lỗi HTTP server: %v", err)
	}

	<-shutdownChan // Đợi server tắt hẳn trước khi thoát
}
