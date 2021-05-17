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
		go streamSingleLectureSource(req.StreamName, sourceName, sourceUrl, req.StreamEnd, req.ID, req.Upload, req.ID)
	}
}

func streamSingleLectureSource(StreamName string, SourceName string, SourceUrl string, streamEnd time.Time, streamID string, uploadRec bool, StreamID string) {
	Workload += 2
	Status = fmt.Sprintf("Streaming %v until %v", StreamName, streamEnd)
	streamEnd = streamEnd.Add(time.Minute * 10)
	go ping()
	for time.Now().Before(streamEnd) {
		log.Println("starting stream")
		cmd := exec.Command(
			"ffmpeg", "-nostats", "-rtsp_transport", "tcp",
			"-stimeout", fmt.Sprintf("%v", streamEnd.Sub(time.Now()).Microseconds()),
			"-t", fmt.Sprintf("%.0f", streamEnd.Sub(time.Now()).Seconds()), // timeout ffmpeg when stream is finished
			"-i", fmt.Sprintf("rtsp://%s", SourceUrl),
			"-map", "0", "-c", "copy", "-f", "mpegts", "-", "-c:v", "libx264", "-preset", "veryfast", "-maxrate", "1500k", "-bufsize", "3000k", "-g", "50", "-r", "25", "-c:a", "aac", "-ar", "44100", "-b:a", "128k",
			"-f", "flv", fmt.Sprintf("%s%s%s", Cfg.IngestBase, StreamName, SourceName))
		log.Println(cmd.String())
		outfile, err := os.OpenFile(fmt.Sprintf("/recordings/vod/%v%v.ts", StreamName, SourceName), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Can't write to disk! %v", err)
			break
		}
		cmd.Stdout = outfile
		err = cmd.Start()
		if err != nil {
			log.Printf("error while processing: %v\n", err)
			continue
		}
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
	Workload -= 2
	Status = "idle"
	ping()
	notifyStreamEnd(streamID)
	convert(fmt.Sprintf("/recordings/vod/%v%v.ts", StreamName, SourceName), fmt.Sprintf("/srv/cephfs/livestream/rec/TUM-Live/vod/%s%s.mp4", StreamName, SourceName))
	if uploadRec {
		upload(fmt.Sprintf("/srv/cephfs/livestream/rec/TUM-Live/vod/%s%s.mp4", StreamName, SourceName), StreamID, SourceName)
	}
}

func notifyStreamEnd(id string) {
	_, err := http.Post(fmt.Sprintf("%s/api/worker/notifyLiveEnd/%s/%v", Cfg.MainBase, Cfg.WorkerID, id), "application.json", bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Printf("couldn't notify server about stream end: %v\n", err)
	}
}

type streamLectureHallRequest struct {
	Sources    map[string]string `json:"sources"` //CAM->123.4.5.6/extron5
	StreamEnd  time.Time         `json:"streamEnd"`
	StreamName string            `json:"streamName"`
	ID         string            `json:"id"`
	Upload     bool              `json:"upload"`
}

type notifyLiveRequest struct {
	StreamID string `json:"streamID"`
	URL      string `json:"url"`     // eg. https://live.lrz.de/livetum/stream/playlist.m3u8
	Version  string `json:"version"` //eg. COMB
}
