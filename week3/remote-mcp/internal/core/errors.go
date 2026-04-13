package core

import "errors"

var (
	ErrLocationNotFound = errors.New("could not find a weather location for the requested input")
	ErrEmptyResult      = errors.New("weather provider returned an empty result")
	ErrRateLimited      = errors.New("weather provider rate limit reached")
	ErrUpstreamTimeout  = errors.New("weather provider request timed out")
	ErrUpstreamFailed   = errors.New("weather provider request failed")
	ErrInvalidUnits     = errors.New("units must be metric or imperial")
	ErrInvalidDays      = errors.New("days must be between 1 and 7")
	ErrLocationRequired = errors.New("location is required")
)
