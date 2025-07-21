package twitch

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const DefaultBaseURL = "https://api.twitch.tv/helix"

type ApiError struct {
	StatusCode int
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("Error returned from twitch API - Status Code: %d", e.StatusCode)
}

type IClient interface {
	get(string, url.Values) (*http.Response, error)
}

type Client struct {
	Log                slog.Logger
	BaseURL            string
	clientId           string
	clientSecret       string
	bearerToken        string
	refreshBearerToken bool
}

func BuildClient(log slog.Logger) *Client {
	clientId, _ := os.LookupEnv("TWITCH_CLIENT_ID")
	clientSecret, _ := os.LookupEnv("TWITCH_CLIENT_SECRET")

	log.Debug("Initialising Twitch Client", "clientIdLength", len(clientId), "clientSecretLength", len(clientSecret))

	return &Client{
		Log:                log,
		BaseURL:            DefaultBaseURL,
		clientId:           clientId,
		clientSecret:       clientSecret,
		refreshBearerToken: true,
	}
}

func (twitch *Client) get(path string, params url.Values) (*http.Response, error) {
	return twitch.makeRequestWithAuth(http.MethodGet, path, params)
}

func (twitch *Client) makeRequestWithAuth(method string, path string, params url.Values) (*http.Response, error) {
	fullUrl := fmt.Sprintf("%s/%s", twitch.BaseURL, strings.TrimPrefix(path, "/"))
	req, err := http.NewRequest(method, fullUrl, nil)
	if err != nil {
		return nil, err
	}
	auth, err := twitch.getAuth()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-Id", auth.id)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.bearer))

	req.URL.RawQuery = params.Encode()

	twitch.Log.Debug("Making Request", "method", method, "url", req.URL.String())

	response, err := http.DefaultClient.Do(req)

	if err == nil && response.StatusCode == 401 {
		// TODO handle credential refresh properly and retry
		twitch.Log.Warn("Got 401 response - invalidating bearer token")
		twitch.refreshBearerToken = true
	}

	return response, err
}
