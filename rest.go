package main

import (
	"TUM-Live-Worker/silencedetect"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func configRouter() {
	server := gin.Default()
	server.GET("/", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})
	// Middleware to reject all requests that don't contain the valid workerID.
	mainGroup := server.Group("/:workerID/")
	mainGroup.Use(func(c *gin.Context) {
		workerID := c.Param("workerID")
		if workerID != Cfg.WorkerID {
			c.AbortWithStatus(http.StatusForbidden)
		}
	})
	mainGroup.POST("/streamLectureHall", streamLectureHall)
	mainGroup.POST("/detectSilence", detectSilence)
	err := server.RunTLS(":443", Cfg.Cert, Cfg.Key)
	if err != nil {
		panic(err)
	}
}

type DetectSilenceReq struct {
	Filename string `json:"filename"`
	StreamID string `json:"stream_id"`
}

func detectSilence(c *gin.Context) {
	var req DetectSilenceReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	go func() {
		sd := silencedetect.New(req.Filename)
		err = sd.ParseSilence()
		if err != nil {
			log.Printf("%v", err)
		} else {
			notifySilenceDetectionResults(sd.Silences, req.StreamID)
		}
	}()
}
