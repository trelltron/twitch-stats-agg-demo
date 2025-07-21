package routes

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/trelltron/twitch-stats-agg-demo/services/twitch"
)

// Test route
type MockTwitchService struct {
	stack  []string
	videos []twitch.Video
	err    error
}

func (m *MockTwitchService) GetUserVideos(clientId string, limit int) ([]twitch.Video, error) {
	m.stack = append(m.stack, fmt.Sprintf("GetUserVideos-%s-%d", clientId, limit))
	return m.videos, m.err
}

func mockService(videos []twitch.Video, err error) MockTwitchService {
	return MockTwitchService{videos: videos, err: err}
}

func errResponse(response *httptest.ResponseRecorder) ErrorResponseBody {
	data := ErrorResponseBody{}
	json.NewDecoder(response.Body).Decode(&data)
	return data
}

func TestRouteNoParams(t *testing.T) {
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)

	service := mockService([]twitch.Video{}, nil)

	RouteGetStreamerStats(c, *slog.Default(), &service)

	err := errResponse(response)

	if !(response.Code == 400 && len(err.Errors) == 2) {
		t.Errorf(`Route test failed - Status %d (expected 400) | Body %v`, response.Code, err)
	}
}

func TestRouteNoLimit(t *testing.T) {
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Params = append(c.Params, gin.Param{Key: "channelId", Value: "testchannel"})

	service := mockService([]twitch.Video{}, nil)

	RouteGetStreamerStats(c, *slog.Default(), &service)

	err := errResponse(response)

	if !(response.Code == 400 && len(err.Errors) == 1) {
		t.Errorf(`Route test failed - Status %d (expected 400) | Body %v`, response.Code, err)
	}
}

func TestRouteTwitchError(t *testing.T) {
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Params = append(c.Params, gin.Param{Key: "channelId", Value: "testchannel"})
	c.Request = httptest.NewRequest("GET", "localhost:3000/streamer/testchannel/stats?limit=100", nil)

	service := mockService([]twitch.Video{}, &twitch.ApiError{})

	RouteGetStreamerStats(c, *slog.Default(), &service)

	err := errResponse(response)

	if !(response.Code == 500 && len(err.Errors) == 1 && len(service.stack) == 1 && service.stack[0] == "GetUserVideos-testchannel-100") {
		t.Errorf(`Route test failed - Status %d (expected 500) | Body %v`, response.Code, err)
	}
}

func TestRouteNoVideos(t *testing.T) {
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Params = append(c.Params, gin.Param{Key: "channelId", Value: "testchannel"})
	c.Request = httptest.NewRequest("GET", "localhost:3000/streamer/testchannel/stats?limit=100", nil)

	service := mockService([]twitch.Video{}, nil)

	RouteGetStreamerStats(c, *slog.Default(), &service)

	err := errResponse(response)

	if !(response.Code == 404 && len(err.Errors) == 1 && len(service.stack) == 1 && service.stack[0] == "GetUserVideos-testchannel-100") {
		t.Errorf(`Route test failed - Status %d (expected 404) | Body %v`, response.Code, err)
	}
}

func TestRouteSuccess(t *testing.T) {
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Params = append(c.Params, gin.Param{Key: "channelId", Value: "testchannel"})
	c.Request = httptest.NewRequest("GET", "localhost:3000/streamer/testchannel/stats?limit=10", nil)

	service := mockService([]twitch.Video{
		{Title: "Title 1", Views: 500, Duration: "2m1s"},
		{Title: "Title 2", Views: 450, Duration: "3m1s"},
		{Title: "Title 3", Views: 480, Duration: "1m1s"},
		{Title: "Title 4", Views: 580, Duration: "3m59s"},
		{Title: "Title 5", Views: 601, Duration: "2m22s"},
		{Title: "Title 6", Views: 572, Duration: "1m11s"},
		{Title: "Title 7", Views: 444, Duration: "3m33s"},
		{Title: "Title 8", Views: 499, Duration: "3m10s"},
		{Title: "Title 9", Views: 654, Duration: "59s"},
		{Title: "Title 10", Views: 399, Duration: "2m"}}, nil)

	RouteGetStreamerStats(c, *slog.Default(), &service)

	if !(response.Code == 200 && len(service.stack) == 1 && service.stack[0] == "GetUserVideos-testchannel-10") {
		t.Errorf(`Route test failed - Status %d (expected 200) | Body %v`, response.Code, response.Body.String())
	}
}

//
// Test generateStats(videos)
//

const FloatComparisonPrecision = 0.01

func floatCompare(a float64, b float64) bool {
	return math.Abs(a-b) < FloatComparisonPrecision
}

func compareStats(t *testing.T, expected Stats, reality Stats) {
	if !(reality.TotalViews == expected.TotalViews &&
		reality.MeanViews == expected.MeanViews &&
		reality.TotalLength == expected.TotalLength &&
		floatCompare(reality.ViewsPerMinute, expected.ViewsPerMinute) &&
		reality.MostViewedVideo.Title == expected.MostViewedVideo.Title &&
		reality.MostViewedVideo.Views == expected.MostViewedVideo.Views) {
		t.Errorf(`generateStats(videos) should return %+v but returns %+v`, expected, reality)
	}
}

func expected(TotalViews int, TotalLength int, MeanViews int, ViewsPerMinute float64, Title string, Views int) Stats {
	MostViewedVideo := SimpleVideo{Title, Views}
	return Stats{
		TotalViews,
		MeanViews,
		TotalLength,
		ViewsPerMinute,
		MostViewedVideo,
	}
}

func TestGenerateStatsSingleVideo(t *testing.T) {
	videos := []twitch.Video{
		{Title: "Title 1", Views: 100, Duration: "1m1s"},
	}
	result := generateStats(videos)

	compareStats(t, expected(100, 61, 100, 98.3606557377, "Title 1", 100), result)
}

func TestGenerateStatsTenSameVideos(t *testing.T) {
	videos := []twitch.Video{
		{Title: "Title 1", Views: 100, Duration: "1m1s"},
		{Title: "Title 2", Views: 100, Duration: "1m1s"},
		{Title: "Title 3", Views: 100, Duration: "1m1s"},
		{Title: "Title 4", Views: 100, Duration: "1m1s"},
		{Title: "Title 5", Views: 100, Duration: "1m1s"},
		{Title: "Title 6", Views: 100, Duration: "1m1s"},
		{Title: "Title 7", Views: 100, Duration: "1m1s"},
		{Title: "Title 8", Views: 100, Duration: "1m1s"},
		{Title: "Title 9", Views: 100, Duration: "1m1s"},
		{Title: "Title 10", Views: 100, Duration: "1m1s"},
	}
	result := generateStats(videos)

	compareStats(t, expected(1000, 610, 100, 98.3606557377, "Title 1", 100), result)
}

func TestGenerateStatsTenSimilarVideos(t *testing.T) {
	videos := []twitch.Video{
		{Title: "Title 1", Views: 500, Duration: "2m1s"},
		{Title: "Title 2", Views: 450, Duration: "3m1s"},
		{Title: "Title 3", Views: 480, Duration: "1m1s"},
		{Title: "Title 4", Views: 580, Duration: "3m59s"},
		{Title: "Title 5", Views: 601, Duration: "2m22s"},
		{Title: "Title 6", Views: 572, Duration: "1m11s"},
		{Title: "Title 7", Views: 444, Duration: "3m33s"},
		{Title: "Title 8", Views: 499, Duration: "3m10s"},
		{Title: "Title 9", Views: 654, Duration: "59s"},
		{Title: "Title 10", Views: 399, Duration: "2m"},
	}
	result := generateStats(videos)

	compareStats(t, expected(5179, 1397, 517, 222.433786686, "Title 9", 654), result)
}

func TestGenerateStatsOutliers(t *testing.T) {
	videos := []twitch.Video{
		{Title: "Title 1", Views: 500, Duration: "2m1s"},
		{Title: "Title 2", Views: 450, Duration: "3m1s"},
		{Title: "Title 3", Views: 8000, Duration: "1h53m12s"},
		{Title: "Title 4", Views: 580, Duration: "3m59s"},
		{Title: "Title 5", Views: 601, Duration: "2m22s"},
		{Title: "Title 6", Views: 764982, Duration: "12m"},
		{Title: "Title 7", Views: 444, Duration: "3m33s"},
		{Title: "Title 8", Views: 499, Duration: "3m10s"},
		{Title: "Title 9", Views: 654, Duration: "59s"},
		{Title: "Title 10", Views: 399, Duration: "2m"},
	}
	result := generateStats(videos)

	compareStats(t, expected(777109, 8777, 77710, 5312.3550188, "Title 6", 764982), result)
}
