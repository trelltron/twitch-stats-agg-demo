package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/trelltron/twitch-stats-agg-demo/services"
)

func BuildRouter(services *services.Services) *gin.Engine {
	router := gin.Default()
	router.GET("/streamer/:channelId/stats", func(c *gin.Context) {
		RouteGetStreamerStats(c, services.Log, &services.Twitch)
	})
	return router
}
