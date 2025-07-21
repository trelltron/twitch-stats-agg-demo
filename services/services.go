package services

import (
	"log/slog"
	"os"

	"github.com/trelltron/twitch-stats-agg-demo/services/twitch"
)

type Services struct {
	Log    slog.Logger
	Twitch twitch.Service
}

func BuildServices() Services {
	log := BuildLogger()
	log.Debug("Logger Initialised")
	twitch := twitch.BuildService(log)
	return Services{log, twitch}
}

func BuildLogger() slog.Logger {
	options := getOptions()

	if useJsonHandler() {
		return *slog.New(slog.NewJSONHandler(os.Stdout, options))
	}
	return *slog.New(slog.NewTextHandler(os.Stdout, options))
}

func getOptions() *slog.HandlerOptions {
	level := getLogLevel()

	return &slog.HandlerOptions{Level: level}
}

func getLogLevel() slog.Level {
	switch LogLevel, _ := os.LookupEnv("LOG_LEVEL"); LogLevel {
	case "ERROR":
		return slog.LevelError
	case "WARN":
		return slog.LevelWarn
	case "INFO":
		return slog.LevelInfo
	default:
		return slog.LevelDebug
	}
}

func useJsonHandler() bool {
	switch UseJSON, _ := os.LookupEnv("JSON_LOGGING"); UseJSON {
	case "TRUE", "true", "True", "T", "t":
		return true
	default:
		return false
	}
}
