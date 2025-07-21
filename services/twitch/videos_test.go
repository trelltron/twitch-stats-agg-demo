package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

type MockClient struct {
	stack  []string
	status int
	videos []Video
	err    error
}

func (m *MockClient) get(path string, params url.Values) (*http.Response, error) {
	m.stack = append(m.stack, fmt.Sprintf("get-%s-%v", path, params))

	if m.err != nil {
		return nil, m.err
	}
	i := 0
	cursor := params.Get("after")
	if len(cursor) > 0 {
		i, _ = strconv.Atoi(cursor)
	}

	body := ResponseBody{}
	if len(m.videos) <= i {
		// fmt.Printf("building empty response")
		return buildEmptyResponse(m.status), nil
	}

	if len(m.videos) > i+100 {
		body.Data = m.videos[i : i+100]
		body.Pagination.Cursor = Cursor(strconv.Itoa(i + 100))
		// fmt.Printf("building response from %+v", body)
		return buildResponse(m.status, body), nil
	}

	body.Data = m.videos[i:]
	// fmt.Printf("building response from %+v", body)
	return buildResponse(m.status, body), nil
}

func setup(e error, status int, v []Video) (Service, *MockClient) {
	c := &MockClient{err: e, status: status, videos: v}
	twitch := Service{
		Log:    *slog.Default(),
		client: c,
	}
	return twitch, c
}

func buildResponse(status int, body ResponseBody) *http.Response {
	r := http.Response{}
	r.StatusCode = status

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(body)
	r.Body = io.NopCloser(buffer)

	return &r
}

func buildEmptyResponse(status int) *http.Response {
	r := http.Response{}
	r.StatusCode = status

	buffer := bytes.NewBufferString("")
	r.Body = io.NopCloser(buffer)

	return &r
}

func generateVideos(n int) []Video {
	var v []Video
	for i := range n {
		v = append(v, Video{Title: fmt.Sprintf("Title %d", i), Views: 100 + i, Duration: "1m1s"})
	}
	return v
}

func TestGetUserVideosError(t *testing.T) {
	twitch, c := setup(&ApiError{}, 500, nil)
	result, err := twitch.GetUserVideos("test", 10)

	if !(len(c.stack) == 1 && len(result) == 0 && err != nil) {
		t.Errorf(`TestGetUserVideosError failed - should have errored`)
	}
}

func TestGetUserVideosNonSuccess(t *testing.T) {
	twitch, c := setup(nil, 400, []Video{})
	result, err := twitch.GetUserVideos("test", 10)

	if !(len(c.stack) == 1 && len(result) == 0 && err != nil) {
		t.Errorf(`TestGetUserVideosNonSuccess failed - should have errored`)
	}
}

func TestGetUserVideosOnePage(t *testing.T) {
	twitch, c := setup(nil, 200, generateVideos(10))
	result, err := twitch.GetUserVideos("test", 10)

	if !(len(c.stack) == 1 && len(result) == 10 && err == nil) {
		t.Errorf(`TestGetUserVideosOnePage failed - stack: %v | len(results): %d | err: %v`, c.stack, len(result), err)
	}
}

func TestGetUserVideosTwoPages(t *testing.T) {
	twitch, c := setup(nil, 200, generateVideos(150))
	result, err := twitch.GetUserVideos("test", 150)

	if !(len(c.stack) == 2 && len(result) == 150 && err == nil) {
		t.Errorf(`TestGetUserVideosTwoPages failed - stack: %v | len(results): %d | err: %v`, c.stack, len(result), err)
	}
}
