package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"remote-mcp/internal/core"
)

type Deps struct {
	Weather *core.WeatherService
}

func New(deps Deps) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "weather-mcp-go",
		Version: "1.0.0",
	}, nil)

	addWeatherTools(server, deps)
	return server
}

func addWeatherTools(server *mcp.Server, deps Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_current_weather",
		Description: "Return current weather details for a location.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in core.CurrentWeatherRequest) (*mcp.CallToolResult, core.CurrentWeatherToolOutput, error) {
		out, err := deps.Weather.GetCurrentWeather(ctx, in.Location, in.Units)
		if err != nil {
			return errorResult(err), core.CurrentWeatherToolOutput{}, nil
		}
		return textResult(out.Summary), out, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_weather_forecast",
		Description: "Return a multi-day weather forecast for a location.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in core.ForecastRequest) (*mcp.CallToolResult, core.ForecastToolOutput, error) {
		out, err := deps.Weather.GetForecast(ctx, in.Location, in.Days, in.Units)
		if err != nil {
			return errorResult(err), core.ForecastToolOutput{}, nil
		}
		return textResult(out.Summary), out, nil
	})
}

func textResult(summary string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: summary},
		},
	}
}

func errorResult(err error) *mcp.CallToolResult {
	payload := map[string]string{
		"error": err.Error(),
	}
	bytes, marshalErr := json.Marshal(payload)
	text := err.Error()
	if marshalErr == nil {
		text = string(bytes)
	}
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}
