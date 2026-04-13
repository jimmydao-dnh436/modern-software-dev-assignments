package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OpenMeteoClient struct {
	settings Settings
	client   *http.Client
}

func NewOpenMeteoClient(settings Settings) *OpenMeteoClient {
	return &OpenMeteoClient{
		settings: settings,
		client: &http.Client{
			Timeout: settings.Timeout,
		},
	}
}

type geocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Country   string  `json:"country"`
		Timezone  string  `json:"timezone"`
		Admin1    string  `json:"admin1"`
	} `json:"results"`
}

type forecastPayload struct {
	Current struct {
		Time                string  `json:"time"`
		Temperature2M       float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		WeatherCode         int     `json:"weather_code"`
		WindSpeed10M        float64 `json:"wind_speed_10m"`
		WindDirection10M    float64 `json:"wind_direction_10m"`
		IsDay               int     `json:"is_day"`
	} `json:"current"`
	Daily struct {
		Time             []string  `json:"time"`
		WeatherCode      []int     `json:"weather_code"`
		Temperature2MMax []float64 `json:"temperature_2m_max"`
		Temperature2MMin []float64 `json:"temperature_2m_min"`
		PrecipitationSum []float64 `json:"precipitation_sum"`
	} `json:"daily"`
}

func (c *OpenMeteoClient) ResolveLocation(ctx context.Context, location string) (ResolvedLocation, error) {
	endpoint := strings.TrimRight(c.settings.WeatherGeocodingBaseURL, "/") + "/search"
	params := url.Values{}
	params.Set("name", location)
	params.Set("count", "1")
	params.Set("language", "en")
	params.Set("format", "json")

	var payload geocodingResponse
	if err := c.getJSON(ctx, endpoint, params, &payload); err != nil {
		return ResolvedLocation{}, err
	}
	if len(payload.Results) == 0 {
		return ResolvedLocation{}, fmt.Errorf("%w: %s", ErrLocationNotFound, location)
	}

	first := payload.Results[0]
	timezone := first.Timezone
	if timezone == "" {
		timezone = "auto"
	}

	return ResolvedLocation{
		Name:      first.Name,
		Latitude:  first.Latitude,
		Longitude: first.Longitude,
		Country:   first.Country,
		Timezone:  timezone,
		Admin1:    first.Admin1,
	}, nil
}

func (c *OpenMeteoClient) GetForecastPayload(
	ctx context.Context,
	latitude float64,
	longitude float64,
	days int,
	temperatureUnit string,
	windSpeedUnit string,
	precipitationUnit string,
	timezone string,
) (forecastPayload, error) {
	endpoint := strings.TrimRight(c.settings.WeatherBaseURL, "/") + "/forecast"
	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%f", latitude))
	params.Set("longitude", fmt.Sprintf("%f", longitude))
	params.Set("current", "temperature_2m,apparent_temperature,weather_code,wind_speed_10m,wind_direction_10m,is_day")
	params.Set("daily", "weather_code,temperature_2m_max,temperature_2m_min,precipitation_sum")
	params.Set("forecast_days", fmt.Sprintf("%d", days))
	params.Set("timezone", timezone)
	params.Set("temperature_unit", temperatureUnit)
	params.Set("wind_speed_unit", windSpeedUnit)
	params.Set("precipitation_unit", precipitationUnit)

	var payload forecastPayload
	if err := c.getJSON(ctx, endpoint, params, &payload); err != nil {
		return forecastPayload{}, err
	}

	if payload.Current.Time == "" && len(payload.Daily.Time) == 0 {
		return forecastPayload{}, ErrEmptyResult
	}
	return payload, nil
}

func (c *OpenMeteoClient) getJSON(ctx context.Context, endpoint string, params url.Values, out any) error {
	attempts := c.settings.RetryAttempts
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", c.settings.UserAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = mapRequestError(err)
		} else {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				lastErr = fmt.Errorf("%w: %v", ErrUpstreamFailed, readErr)
			} else if resp.StatusCode == http.StatusTooManyRequests {
				lastErr = ErrRateLimited
			} else if resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("%w: upstream returned %d", ErrUpstreamFailed, resp.StatusCode)
			} else if resp.StatusCode >= 400 {
				lastErr = fmt.Errorf("%w: upstream returned %d", ErrUpstreamFailed, resp.StatusCode)
			} else if err := json.Unmarshal(body, out); err != nil {
				lastErr = fmt.Errorf("%w: could not decode upstream JSON: %v", ErrUpstreamFailed, err)
			} else {
				return nil
			}
		}

		if attempt+1 < attempts {
			backoff := c.settings.Backoff * time.Duration(1<<attempt)
			time.Sleep(backoff)
		}
	}

	return lastErr
}

func mapRequestError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrUpstreamTimeout
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return ErrUpstreamTimeout
	}

	return fmt.Errorf("%w: %v", ErrUpstreamFailed, err)
}
