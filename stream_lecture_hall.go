package main

import (
	"TUM-Live-Worker/silencedetect"
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
		go streamSingleLectureSource(sourceName, sourceUrl, req.StreamEnd, req.ID, req.Upload, req.ID, req)
	}
}

func streamSingleLectureSource(SourceName string, SourceUrl string, streamEnd time.Time, streamID string, uploadRec bool, StreamID string, req streamLectureHallRequest) {
	Workload += 2
	Status = fmt.Sprintf("Streaming %v until %v", req.StreamName, streamEnd)
	streamEnd = streamEnd.Add(time.Minute * 10)
	go ping()
	for time.Now().Before(streamEnd) {
		log.Println("starting stream")
		cmd := exec.Command(
			"ffmpeg", "-nostats", "-rtsp_transport", "tcp",
			"-t", fmt.Sprintf("%.0f", streamEnd.Sub(time.Now()).Seconds()), // timeout ffmpeg when stream is finished
			"-i", fmt.Sprintf("rtsp://%s", SourceUrl),
			"-map", "0", "-c", "copy", "-f", "mpegts", "-", "-c:v", "libx264", "-preset", "veryfast", "-maxrate", "1500k", "-bufsize", "3000k", "-g", "50", "-r", "25", "-c:a", "aac", "-ar", "44100", "-b:a", "128k",
			"-f", "flv", fmt.Sprintf("%s%s%s", Cfg.IngestBase, req.StreamName, SourceName))
		log.Println(cmd.String())
		outfile, err := os.OpenFile(fmt.Sprintf("/recordings/vod/%v%v.ts", req.StreamName, SourceName), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
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
			URL:      fmt.Sprintf("https://live.stream.lrz.de/livetum/%v%v/playlist.m3u8", req.StreamName, SourceName),
			Version:  SourceName,
		})
		_, err = http.Post(fmt.Sprintf("%v/api/worker/notifyLive/%v", Cfg.MainBase, Cfg.WorkerID), "application/json", bytes.NewBuffer(notifyLiveBody))
		if err != nil {
			log.Printf("Error notifying server: %v\n", err)
		}
		err = cmd.Wait()
		if err != nil {
			log.Printf("Error while waiting: %v\n", err)
			delete(streamJobs, fmt.Sprintf("%s%s", req.StreamName, SourceName))
			err = outfile.Close()
			if err != nil {
				log.Printf("Couldn't close outfile: %s\n", err)
			}
			continue
		}
		delete(streamJobs, fmt.Sprintf("%s%s", req.StreamName, SourceName))
		err = outfile.Close()
		if err != nil {
			log.Printf("Couldn't close outfile: %s\n", err)
		}
	}
	log.Printf("finished streaming %v%v", req.StreamName, SourceName)
	Workload -= 2
	Status = "idle"
	ping()
	notifyStreamEnd(streamID)
	targetFolder := fmt.Sprintf("/srv/cephfs/livestream/rec/TUM-Live/%d/%s/%s/%s", req.Semester, req.TeachingTerm, req.Slug, req.StreamName)
	err := os.MkdirAll(targetFolder, 0750)
	if err != nil {
		log.Printf("Could not create target folder: %v", err)
		return
	}
	targetFile := fmt.Sprintf("%s/%s%s.mp4", targetFolder, req.StreamName, SourceName)
	convert(fmt.Sprintf("/recordings/vod/%v%v.ts", req.StreamName, SourceName), targetFile)
	if uploadRec {
		upload(targetFile, StreamID, SourceName)
	}
	if len(req.Sources) == 1 || SourceName == "PRES" {
		sd := silencedetect.New(targetFile)
		err := sd.ParseSilence()
		if err != nil {
			log.Printf("%v", err)
		} else {
			notifySilenceDetectionResults(sd.Silences, streamID)
		}
	}
}

func notifyStreamEnd(id string) {
	_, err := http.Post(fmt.Sprintf("%s/api/worker/notifyLiveEnd/%s/%v", Cfg.MainBase, Cfg.WorkerID, id), "application.json", bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Printf("couldn't notify server about stream end: %v\n", err)
	}
}

type streamLectureHallRequest struct {
	Sources      map[string]string `json:"sources"` //CAM->123.4.5.6/extron5
	StreamEnd    time.Time         `json:"streamEnd"`
	StreamName   string            `json:"streamName"`
	ID           string            `json:"id"`
	Upload       bool              `json:"upload"`
	Semester     int               `json:"semester"`
	TeachingTerm string            `json:"teachingTerm"`
	Slug         string            `json:"slug"`
}

type notifyLiveRequest struct {
	StreamID string `json:"streamID"`
	URL      string `json:"url"`     // eg. https://live.lrz.de/livetum/stream/playlist.m3u8
	Version  string `json:"version"` //eg. COMB
}
