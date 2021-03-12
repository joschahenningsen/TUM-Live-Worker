package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func convertAndUpload() {
	if Busy {
		return
	}
	res, err := http.Get("http://backend:7002/api/worker/getJobs/" + Cfg.WorkerID)
	if err != nil {
		log.Printf("couldn't get jobs%v\n", err)
		return
	}
	jsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("couldn't get jobs%v\n", err)
		return
	}
	var job jobData
	err = json.Unmarshal(jsonData, &job)
	if err != nil {
		log.Printf("couldn't parse job%v\n", err)
		return
	}
	Busy = true

	convert(job.Path)
	newPath := strings.Replace(job.Path, ".flv", ".mp4", 1)
	upload(newPath)

}

func upload(path string) {
	log.Printf("Uploading %v", path)
	//http.PostForm(Cfg.LrzUploadURL, url.Values{})
	// todo: use http instead of curl
	cmd := exec.Command(
		"curl",
		"-F", fmt.Sprintf("benutzer=%v", Cfg.LrzUser),
		"-F", fmt.Sprintf("mailadresse=%v", Cfg.LrzMail),
		"-F", fmt.Sprintf("telefon=%v", Cfg.LrzPhone),
		"-F", "unidir=tum",
		"-F", fmt.Sprintf("subdir=%v", Cfg.LrzSubDir),
		"-F", fmt.Sprintf("info=%v", "testupload"),
		"-F", fmt.Sprintf("filename=%v", path),
		Cfg.LrzUploadURL,
	)
	println(cmd.String())
	err := cmd.Start()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
}

func convert(file string) {
	newPath := strings.Replace(file, ".flv", ".mp4", 1)
	log.Printf("Processing %v output: %v", file, newPath)
	cmd := exec.Command("ffmpeg", "-i", file, "-c:v", "libx264", "-crf", "19", "-strict", "experimental", newPath)
	err := cmd.Start()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	_ = os.Remove(file)
}

type jobData struct {
	Id   uint   `json:"id"`
	Path string `json:"path"`
}
