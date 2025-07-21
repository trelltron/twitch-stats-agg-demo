package twitch

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestTwitchClientMakeRequestWithAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !(r.Method == "GET" &&
			r.URL.Path == "/testpath" &&
			r.URL.Query().Get("key1") == "val1" &&
			r.Header.Get("Client-Id") == "client-id" &&
			r.Header.Get("Authorization") == "Bearer imnotabear") {
			t.Errorf(`TestTwitchClientMakeRequestWithAuth failed - %s %s?%s Headers: %v`, r.Method, r.URL.Path, r.URL.Query().Encode(), r.Header)
		}
	})
	s := httptest.NewServer(handler)
	defer s.Close()

	client := Client{
		Log:                *slog.Default(),
		BaseURL:            s.URL,
		clientId:           "client-id",
		clientSecret:       "itsasecret",
		bearerToken:        "imnotabear",
		refreshBearerToken: false,
	}

	res, err := client.makeRequestWithAuth("GET", "testpath", url.Values{"key1": {"val1"}})

	if !(res.StatusCode == 200 && err == nil) {
		t.Errorf(`TestTwitchClientMakeRequestWithAuth failed - Status: %d | err %v`, res.StatusCode, err)
	}
}
