package twitch

import (
	"log/slog"
)

type Service struct {
	Log    slog.Logger
	client IClient
}

func BuildService(log slog.Logger) Service {
	return Service{
		Log:    log,
		client: BuildClient(log),
	}

}
