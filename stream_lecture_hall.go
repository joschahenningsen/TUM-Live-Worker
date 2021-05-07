package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var streamJobs = map[string]*os.Process{}

func streamLectureHall(context *gin.Context) {
	body, _ := ioutil.ReadAll(context.Request.Body)
	var req streamLectureHallRequest
	err := json.Unmarshal(body, &req)
	if err != nil {
		log.Println("invalid request for streamLectureHall")
		context.AbortWithStatus(http.StatusBadRequest)
		return
	}
	go stream(req)
}

func stream(req streamLectureHallRequest) {
	for sourceName, sourceUrl := range req.Sources {
		log.Printf("%v:%v\n", sourceName, sourceUrl)
		go streamSingleLectureSource(req.StreamName, sourceName, sourceUrl, req.StreamEnd, req.ID)
	}
}

func streamSingleLectureSource(StreamName string, SourceName string, SourceUrl string, streamEnd time.Time, streamID string) {
	Workload += 2
	defer func() { Workload -= 2 }() // todo possible race condition?
	for streamEnd.After(time.Now()) {
		log.Println("starting stream")
		cmd := exec.Command(
			"ffmpeg", "-nostats", "-rtsp_transport", "tcp",
			"-i", fmt.Sprintf("rtsp://%s", SourceUrl),
			"-t", fmt.Sprintf("%.0f", streamEnd.Sub(time.Now()).Seconds()), // timeout ffmpeg when stream is finished
			"-map", "0", "-c", "copy", "-f", "mpegts", "-", "-c:v", "libx264", "-preset", "veryfast", "-maxrate", "1500k", "-bufsize", "3000k", "-g", "50", "-r", "25", "-c:a", "aac", "-ar", "44100", "-b:a", "128k",
			"-f", "flv", fmt.Sprintf("%s%s%s >> /recordings/vod/%v%v.ts", Cfg.IngestBase, StreamName, SourceName, StreamName, SourceName))
		log.Println(cmd.String())
		err := cmd.Start()
		if err != nil {
			log.Printf("error while processing: %v\n", err)
			continue
		}
		streamJobs[fmt.Sprintf("%s%s", StreamName, SourceName)] = cmd.Process
		log.Println(cmd.Process.Pid)
		notifyLiveBody, _ := json.Marshal(notifyLiveRequest{
			StreamID: streamID,
			URL:      fmt.Sprintf("https://live.stream.lrz.de/livetum/%v%v/playlist.m3u8", StreamName, SourceName),
			Version:  SourceName,
		})
		_, err = http.Post(fmt.Sprintf("%v/api/worker/notifyLive/%v", Cfg.MainBase, Cfg.WorkerID), "application/json", bytes.NewBuffer(notifyLiveBody))
		if err != nil {
			log.Printf("Error notifying server: %v\n", err)
		}
		err = cmd.Wait()
		if err != nil {
			log.Printf("Error while waiting: %v\n", err)
			delete(streamJobs, fmt.Sprintf("%s%s", StreamName, SourceName))
			continue
		}
		delete(streamJobs, fmt.Sprintf("%s%s", StreamName, SourceName))
	}
	log.Printf("finished streaming %v%v", StreamName, SourceName)
}

type streamLectureHallRequest struct {
	Sources    map[string]string `json:"sources"` //CAM->123.4.5.6/extron5
	StreamEnd  time.Time         `json:"streamEnd"`
	StreamName string            `json:"streamName"`
	ID         string            `json:"id"`
}

type notifyLiveRequest struct {
	StreamID string `json:"streamID"`
	URL      string `json:"url"`     // eg. https://live.lrz.de/livetum/stream/playlist.m3u8
	Version  string `json:"version"` //eg. COMB
}
