package routes

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trelltron/twitch-stats-agg-demo/services/twitch"
)

type ITwitch interface {
	GetUserVideos(string, int) ([]twitch.Video, error)
}

type parsedInput struct {
	channelId string
	limit     int
	errors    []string
}

type ErrorResponseBody struct {
	Errors []string `json:"errors"`
}

type SimpleVideo struct {
	Title string `json:"title"`
	Views int    `json:"views"`
}

type Stats struct {
	TotalViews      int         `json:"totalViews"`
	MeanViews       int         `json:"meanViews"`
	TotalLength     int         `json:"totalLength"`
	ViewsPerMinute  float64     `json:"viewsPerMinute"`
	MostViewedVideo SimpleVideo `json:"mostViewedVideo"`
}

func RouteGetStreamerStats(c *gin.Context, log slog.Logger, twitch ITwitch) {

	input := parseInput(c)

	if len(input.errors) > 0 {
		c.JSON(http.StatusBadRequest, ErrorResponseBody{Errors: input.errors})
		return
	}

	result, err := twitch.GetUserVideos(input.channelId, input.limit)

	if err != nil {
		c.JSON(500, ErrorResponseBody{Errors: []string{"Something went wrong"}})
		return
	}

	if len(result) == 0 {
		c.JSON(404, ErrorResponseBody{Errors: []string{"No videos found for this userId"}})
		return
	}

	stats := generateStats(result)

	log.Debug("Returning stats blob", "stats", stats)

	c.JSON(http.StatusOK, stats)

}

func parseInput(c *gin.Context) parsedInput {
	channelId := c.Param("channelId")
	limit, err := strconv.Atoi(c.Query("limit"))

	errors := []string{}
	if len(channelId) == 0 {
		// This state probably shouldn't be possible for the current route definition
		errors = append(errors, "Missing channel ID")
	}
	if err != nil || limit == 0 {
		errors = append(errors, "Missing or invalid limit parameter")
	}

	return parsedInput{channelId, limit, errors}
}

func generateStats(videos []twitch.Video) Stats {
	if len(videos) == 0 {
		// This case should already be handled in the route handler
		return Stats{}
	}
	var (
		totalViews  int     = 0
		totalLength float64 = 0
	)
	mostViewed := SimpleVideo{Title: "", Views: 0}
	for _, video := range videos {
		totalViews = totalViews + video.Views
		duration, e := time.ParseDuration(video.Duration)
		if e == nil {
			// TODO: should handle any errors in this parsing
			totalLength = totalLength + duration.Seconds()
		}
		if video.Views > mostViewed.Views {
			mostViewed = SimpleVideo{Title: video.Title, Views: video.Views}
		}
	}

	return Stats{
		TotalViews:      totalViews,
		MeanViews:       totalViews / len(videos),
		TotalLength:     int(totalLength),
		ViewsPerMinute:  float64(totalViews) * 60 / totalLength,
		MostViewedVideo: mostViewed,
	}
}
