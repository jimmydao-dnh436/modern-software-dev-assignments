# Weather MCP Server (Go, Remote HTTP, No OAuth Yet)

- `internal/core/` giữ business logic và Open-Meteo client
- `internal/mcpserver/` đăng ký 2 weather tools
- `internal/httpapi/` bọc HTTP layer cho remote MCP
- `cmd/mcp-remote/main.go` là entrypoint chạy server

## Cấu trúc thư mục

```text
 go-weather-mcp/
 ├── .env.example
 ├── go.mod
 ├── README.md
 ├── cmd/
 │   └── mcp-remote/
 │       └── main.go
 └── internal/
     ├── core/
     │   ├── client.go
     │   ├── config.go
     │   ├── errors.go
     │   ├── models.go
     │   └── service.go
     ├── httpapi/
     │   └── router.go
     └── mcpserver/
         └── server.go
```

## Hai tool có sẵn

### `get_current_weather`
Input:

```json
{
  "location": "Ho Chi Minh City",
  "units": "metric"
}
```

### `get_weather_forecast`
Input:

```json
{
  "location": "Da Nang",
  "days": 3,
  "units": "metric"
}
```

## Chạy project

```bash
cp .env.example .env
go mod tidy
go run ./cmd/mcp-remote
```

Server mặc định chạy tại:

- MCP endpoint: `http://127.0.0.1:8000/mcp`
- Health check: `http://127.0.0.1:8000/health`

## Test bằng MCP Inspector

Trong MCP Inspector:

- Transport: **Streamable HTTP**
- URL: `http://127.0.0.1:8000/mcp`

Rồi gọi thử:

### Tool 1

```json
{
  "location": "Ho Chi Minh City",
  "units": "metric"
}
```

### Tool 2

```json
{
  "location": "Ha Noi",
  "days": 5,
  "units": "imperial"
}
```

## Ghi chú

- Đây là bản **remote MCP chưa có OAuth**.
- HTTP layer đã có CORS cho Inspector và kiểm tra `Origin` cơ bản.
- Phần JSON-RPC / MCP transport do Go SDK xử lý qua `NewStreamableHTTPHandler(...)`.
