package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func convertAndUpload() {
	if Busy {
		return
	}
	res, err := http.Get("http://localhost:8080/api/worker/getJobs/" + Cfg.WorkerID)
	if err != nil {
		log.Printf("couldn't get jobs%v\n", err)
		return
	}
	if res.StatusCode != 200 {
		log.Println("no job.")
		return
	}
	jsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("couldn't get jobs%v\n", err)
		return
	}
	println(jsonData)
	var job jobData
	err = json.Unmarshal(jsonData, &job)
	if err != nil {
		log.Printf("couldn't parse job%v\n", err)
		return
	}
	Busy = true

	convert(job.Path)
	newPath := strings.Replace(job.Path, ".flv", ".mp4", 1)
	if job.Upload {
		upload(newPath, job)
	}
	Busy = false

}

func upload(path string, job jobData) {
	log.Printf("Uploading %v", path)
	pathparts := strings.Split(path, "/")
	cmd := exec.Command("curl",
		"-F", fmt.Sprintf("'benutzer=%v'", Cfg.LrzUser),
		"-F", fmt.Sprintf("'mailadresse=%v'", Cfg.LrzMail),
		"-F", fmt.Sprintf("'telefon=%v'", Cfg.LrzPhone),
		"-F", "'unidir=tum'",
		"-F", fmt.Sprintf("'subdir=%v'", Cfg.LrzSubDir),
		"-F", fmt.Sprintf("'info='"),
		"-F", fmt.Sprintf("'filename=@%v'", path),
		Cfg.LrzUploadURL)

	err := cmd.Start()
	if err != nil {
		log.Printf("Error starting curl: %v", err)
	}
	err = cmd.Wait()

	if err != nil {
		log.Printf("Error executing curl: %v", err)
	}
	log.Println("Uploaded file to lrz.")
	createVodData := putVodData{
		Name:     job.Name,
		Start:    job.StreamStart,
		HlsUrl:   "https://stream.lrz.de/vod/_definst_/mp4:tum/RBG/" + pathparts[len(pathparts)-1] + "/playlist.m3u8",
		StreamId: job.StreamId,
	}
	send, _ := json.Marshal(createVodData)
	_, err = http.Post("http://backend:8080/api/worker/putVOD/"+Cfg.WorkerID,
		"application/json",
		bytes.NewBuffer(send))
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func convert(file string) {
	newPath := strings.Replace(file, ".flv", ".mp4", 1)
	log.Printf("Processing %v output: %v", file, newPath)
	cmd := exec.Command("ffmpeg", "-i", file, "-c:v", "libx264", "-crf", "19", "-strict", "experimental", newPath)
	err := cmd.Start()
	if err != nil {
		log.Printf("error while processing: %v\n", err)
		return
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Error while waiting: %v\n", err)
		return
	}
	// _ = os.Remove(file)
}

type jobData struct {
	Id          uint      `json:"id"`
	Name        string    `json:"name"`
	StreamId    uint      `json:"streamId"`
	StreamStart time.Time `json:"streamStart"`
	Path        string    `json:"path"`
	Upload      bool
}

type putVodData struct {
	Name     string
	Start    time.Time
	HlsUrl   string
	StreamId uint
}
