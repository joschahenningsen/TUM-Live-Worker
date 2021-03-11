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
	res, err := http.Get("http://backend:7002/api/worker/getJobs/" + WorkerID)
	if err != nil {
		log.Printf("couldn't get jobs%v\n", err)
		return
	}
	jsonData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("couldn't get jobs%v\n", err)
		return
	}
	var job job
	err = json.Unmarshal(jsonData, &job)
	if err != nil {
		log.Printf("couldn't parse job%v\n", err)
		return
	}
	Busy = true
	println(job.path)
	exec.Command(fmt.Sprintf("ffmpeg -i %v -c:v libx264 -crf 19 -strict experimental %v", job.path, strings.Replace(job.path, ".flv", ".mp4", 1)))
	_ = os.Remove(job.path)

}

type job struct {
	id   uint
	path string
}
