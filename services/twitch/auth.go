package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type OAuthResponse struct {
	Token string `json:"access_token"`
}

type OAuthError struct {
	StatusCode int
}

func (e *OAuthError) Error() string {
	return fmt.Sprintf("OAuth grant failed - Status Code: %d", e.StatusCode)
}

type AuthDetails struct {
	id     string
	bearer string
}

func (twitch *Client) getAuth() (AuthDetails, error) {
	if twitch.refreshBearerToken {
		if err := twitch.getNewToken(); err != nil {
			return AuthDetails{}, err
		}
	}
	return AuthDetails{
		id:     twitch.clientId,
		bearer: twitch.bearerToken,
	}, nil
}

func (twitch *Client) getNewToken() error {
	data := url.Values{}
	data.Set("client_id", twitch.clientId)
	data.Set("client_secret", twitch.clientSecret)
	data.Set("grant_type", "client_credentials")
	res, err := http.Post("https://id.twitch.tv/oauth2/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		twitch.Log.Error("OAuth grant failed", "err", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		twitch.Log.Error("OAuth grant failed", "StatusCode", res.StatusCode, "details", string(body))
		return &OAuthError{StatusCode: res.StatusCode}
	}

	var resData OAuthResponse
	if err := json.NewDecoder(res.Body).Decode(&resData); err != nil {
		twitch.Log.Error("OAuth grant failed", "err", err)
		return err
	}
	twitch.bearerToken = resData.Token
	return nil

}
