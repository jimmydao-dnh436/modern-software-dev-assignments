package core

import (
	"context"
	"fmt"
	"strings"
)

var weatherCodes = map[int]string{
	0:  "Clear sky",
	1:  "Mainly clear",
	2:  "Partly cloudy",
	3:  "Overcast",
	45: "Fog",
	48: "Depositing rime fog",
	51: "Light drizzle",
	53: "Moderate drizzle",
	55: "Dense drizzle",
	61: "Light rain",
	63: "Moderate rain",
	65: "Heavy rain",
	71: "Light snow",
	73: "Moderate snow",
	75: "Heavy snow",
	80: "Rain showers",
	81: "Moderate rain showers",
	82: "Violent rain showers",
	95: "Thunderstorm",
}

type WeatherService struct {
	client *OpenMeteoClient
}

func NewWeatherService(client *OpenMeteoClient) *WeatherService {
	return &WeatherService{client: client}
}

func (s *WeatherService) GetCurrentWeather(ctx context.Context, location string, units Units) (CurrentWeatherToolOutput, error) {
	normalizedLocation, err := NormalizeLocation(location)
	if err != nil {
		return CurrentWeatherToolOutput{}, err
	}
	normalizedUnits, err := NormalizeUnits(units)
	if err != nil {
		return CurrentWeatherToolOutput{}, err
	}

	resolved, err := s.client.ResolveLocation(ctx, normalizedLocation)
	if err != nil {
		return CurrentWeatherToolOutput{}, err
	}

	payload, err := s.client.GetForecastPayload(
		ctx,
		resolved.Latitude,
		resolved.Longitude,
		1,
		s.temperatureUnit(normalizedUnits),
		s.windSpeedUnit(normalizedUnits),
		s.precipitationUnit(normalizedUnits),
		resolved.Timezone,
	)
	if err != nil {
		return CurrentWeatherToolOutput{}, err
	}

	current, err := s.buildCurrent(payload)
	if err != nil {
		return CurrentWeatherToolOutput{}, err
	}
	forecast, err := s.buildForecast(payload)
	if err != nil {
		return CurrentWeatherToolOutput{}, err
	}

	summary := fmt.Sprintf(
		"Current weather for %s: %.1f%s, %s, wind %.1f %s.",
		resolved.DisplayName(),
		current.Temperature,
		TemperatureSymbol(normalizedUnits),
		strings.ToLower(current.WeatherDescription),
		current.WindSpeed,
		WindSpeedSymbol(normalizedUnits),
	)

	return CurrentWeatherToolOutput{
		Summary:  summary,
		Location: resolved,
		Units:    normalizedUnits,
		Warnings: []string{},
		Current:  current,
		Forecast: firstForecastOnly(forecast),
	}, nil
}

func (s *WeatherService) GetForecast(ctx context.Context, location string, days int, units Units) (ForecastToolOutput, error) {
	normalizedLocation, err := NormalizeLocation(location)
	if err != nil {
		return ForecastToolOutput{}, err
	}
	if err := ValidateDays(days); err != nil {
		return ForecastToolOutput{}, err
	}
	normalizedUnits, err := NormalizeUnits(units)
	if err != nil {
		return ForecastToolOutput{}, err
	}

	resolved, err := s.client.ResolveLocation(ctx, normalizedLocation)
	if err != nil {
		return ForecastToolOutput{}, err
	}

	payload, err := s.client.GetForecastPayload(
		ctx,
		resolved.Latitude,
		resolved.Longitude,
		days,
		s.temperatureUnit(normalizedUnits),
		s.windSpeedUnit(normalizedUnits),
		s.precipitationUnit(normalizedUnits),
		resolved.Timezone,
	)
	if err != nil {
		return ForecastToolOutput{}, err
	}

	current, err := s.buildCurrent(payload)
	if err != nil {
		return ForecastToolOutput{}, err
	}
	forecast, err := s.buildForecast(payload)
	if err != nil {
		return ForecastToolOutput{}, err
	}
	if len(forecast) == 0 {
		return ForecastToolOutput{}, ErrEmptyResult
	}
	if len(forecast) > days {
		forecast = forecast[:days]
	}

	summary := fmt.Sprintf(
		"%d-day forecast for %s: %s today, high %.1f%s, low %.1f%s, precipitation %.1f%s.",
		days,
		resolved.DisplayName(),
		strings.ToLower(forecast[0].WeatherDescription),
		forecast[0].TemperatureMax,
		TemperatureSymbol(normalizedUnits),
		forecast[0].TemperatureMin,
		TemperatureSymbol(normalizedUnits),
		forecast[0].PrecipitationSum,
		PrecipitationSymbol(normalizedUnits),
	)

	return ForecastToolOutput{
		Summary:  summary,
		Location: resolved,
		Units:    normalizedUnits,
		Warnings: []string{},
		Current:  current,
		Forecast: forecast,
	}, nil
}

func (s *WeatherService) buildCurrent(payload forecastPayload) (CurrentWeather, error) {
	if payload.Current.Time == "" {
		return CurrentWeather{}, ErrEmptyResult
	}

	return CurrentWeather{
		Time:                payload.Current.Time,
		Temperature:         payload.Current.Temperature2M,
		ApparentTemperature: payload.Current.ApparentTemperature,
		WindSpeed:           payload.Current.WindSpeed10M,
		WindDirection:       payload.Current.WindDirection10M,
		WeatherCode:         payload.Current.WeatherCode,
		WeatherDescription:  describeWeatherCode(payload.Current.WeatherCode),
		IsDay:               payload.Current.IsDay == 1,
	}, nil
}

func (s *WeatherService) buildForecast(payload forecastPayload) ([]ForecastDay, error) {
	dates := payload.Daily.Time
	codes := payload.Daily.WeatherCode
	highs := payload.Daily.Temperature2MMax
	lows := payload.Daily.Temperature2MMin
	rain := payload.Daily.PrecipitationSum

	if len(dates) == 0 {
		return nil, ErrEmptyResult
	}

	n := minInt(len(dates), len(codes), len(highs), len(lows), len(rain))
	if n == 0 {
		return nil, ErrEmptyResult
	}

	forecast := make([]ForecastDay, 0, n)
	for i := 0; i < n; i++ {
		forecast = append(forecast, ForecastDay{
			Date:               dates[i],
			TemperatureMax:     highs[i],
			TemperatureMin:     lows[i],
			PrecipitationSum:   rain[i],
			WeatherCode:        codes[i],
			WeatherDescription: describeWeatherCode(codes[i]),
		})
	}
	return forecast, nil
}

func (s *WeatherService) temperatureUnit(units Units) string {
	if units == UnitsImperial {
		return "fahrenheit"
	}
	return "celsius"
}

func (s *WeatherService) windSpeedUnit(units Units) string {
	if units == UnitsImperial {
		return "mph"
	}
	return "kmh"
}

func (s *WeatherService) precipitationUnit(units Units) string {
	if units == UnitsImperial {
		return "inch"
	}
	return "mm"
}

func describeWeatherCode(code int) string {
	if label, ok := weatherCodes[code]; ok {
		return label
	}
	return fmt.Sprintf("Weather code %d", code)
}

func minInt(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
}

func firstForecastOnly(days []ForecastDay) []ForecastDay {
	if len(days) == 0 {
		return nil
	}
	return []ForecastDay{days[0]}
}
