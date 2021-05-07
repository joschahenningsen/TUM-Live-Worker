package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func configRouter() {
	server := gin.Default()
	// Middleware to reject all requests that don't contain the valid workerID.
	mainGroup := server.Group("/:workerID/")
	mainGroup.Use(func(c *gin.Context) {
		workerID := c.Param("workerID")
		if workerID != Cfg.WorkerID {
			c.AbortWithStatus(http.StatusForbidden)
		}
	})
	mainGroup.POST("/streamLectureHall", streamLectureHall)
	err := server.RunTLS(":443", Cfg.Cert, Cfg.Key)
	if err != nil {
		panic(err)
	}
}
