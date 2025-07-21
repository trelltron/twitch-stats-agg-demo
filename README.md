# twitch-stats-agg-demo

## Configuring application

1. Copy `.env.example` to `.env` and set the twitch client ID and secret
2. Add any other environment variables to the `.env` file as necessary

### Environment variables

| Environment variable | Default | Possible values | Description |
| --- | --- | --- | --- |
| `SERVER_ADDRESS` | `localhost:3000` | Valid address | HTTP server listening address |
| `LOG_LEVEL` | `debug` | `debug`, `info`, `warning`, `error` | Logging level |
| `JSON_LOGGING` | `false` | `true` or `false` | set logger to use json output |
| `TWITCH_CLIENT_ID` | | | Twitch Client ID |
| `TWITCH_CLIENT_SECRET` | | | Twitch Client Secret |

## Running application

```
docker-compose up --build -d
```