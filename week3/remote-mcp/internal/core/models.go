package core

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type Units string

const (
	UnitsMetric   Units = "metric"
	UnitsImperial Units = "imperial"
)

type CurrentWeatherRequest struct {
	Location string `json:"location" jsonschema:"human-readable location to resolve"`
	Units    Units  `json:"units,omitempty" jsonschema:"metric or imperial units"`
}

type ForecastRequest struct {
	Location string `json:"location" jsonschema:"human-readable location to resolve"`
	Days     int    `json:"days,omitempty" jsonschema:"number of forecast days between 1 and 7"`
	Units    Units  `json:"units,omitempty" jsonschema:"metric or imperial units"`
}

type ResolvedLocation struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country,omitempty"`
	Timezone  string  `json:"timezone,omitempty"`
	Admin1    string  `json:"admin1,omitempty"`
}

func (r ResolvedLocation) DisplayName() string {
	if r.Country != "" {
		return fmt.Sprintf("%s, %s", r.Name, r.Country)
	}
	return r.Name
}

type CurrentWeather struct {
	Time                string  `json:"time"`
	Temperature         float64 `json:"temperature"`
	ApparentTemperature float64 `json:"apparent_temperature,omitempty"`
	WindSpeed           float64 `json:"wind_speed"`
	WindDirection       float64 `json:"wind_direction,omitempty"`
	WeatherCode         int     `json:"weather_code"`
	WeatherDescription  string  `json:"weather_description"`
	IsDay               bool    `json:"is_day"`
}

type ForecastDay struct {
	Date               string  `json:"date"`
	TemperatureMax     float64 `json:"temperature_max"`
	TemperatureMin     float64 `json:"temperature_min"`
	PrecipitationSum   float64 `json:"precipitation_sum"`
	WeatherCode        int     `json:"weather_code"`
	WeatherDescription string  `json:"weather_description"`
}

type CurrentWeatherToolOutput struct {
	Summary  string           `json:"summary"`
	Location ResolvedLocation `json:"location"`
	Units    Units            `json:"units"`
	Warnings []string         `json:"warnings,omitempty"`
	Current  CurrentWeather   `json:"current"`
	Forecast []ForecastDay    `json:"forecast,omitempty"`
}

type ForecastToolOutput struct {
	Summary  string           `json:"summary"`
	Location ResolvedLocation `json:"location"`
	Units    Units            `json:"units"`
	Warnings []string         `json:"warnings,omitempty"`
	Current  CurrentWeather   `json:"current,omitempty"`
	Forecast []ForecastDay    `json:"forecast"`
}

func NormalizeLocation(value string) (string, error) {
	normalized, err := url.QueryUnescape(strings.TrimSpace(value))
	if err != nil {
		normalized = strings.TrimSpace(value)
	}
	normalized = strings.ReplaceAll(normalized, "+", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")
	normalized = strings.ReplaceAll(normalized, "-", " ")

	camelBoundary := regexp.MustCompile(`([a-z])([A-Z])`)
	normalized = camelBoundary.ReplaceAllString(normalized, `$1 $2`)
	normalized = strings.Trim(normalized, " \t\r\n,;:!?/")
	normalized = strings.Join(strings.Fields(normalized), " ")
	if normalized == "" {
		return "", ErrLocationRequired
	}
	return normalized, nil
}

func NormalizeUnits(value Units) (Units, error) {
	switch Units(strings.ToLower(strings.TrimSpace(string(value)))) {
	case "", UnitsMetric:
		return UnitsMetric, nil
	case UnitsImperial:
		return UnitsImperial, nil
	default:
		return "", ErrInvalidUnits
	}
}

func ValidateDays(days int) error {
	if days < 1 || days > 7 {
		return ErrInvalidDays
	}
	return nil
}

func TemperatureSymbol(units Units) string {
	if units == UnitsImperial {
		return "°F"
	}
	return "°C"
}

func WindSpeedSymbol(units Units) string {
	if units == UnitsImperial {
		return "mph"
	}
	return "km/h"
}

func PrecipitationSymbol(units Units) string {
	if units == UnitsImperial {
		return "in"
	}
	return "mm"
}
