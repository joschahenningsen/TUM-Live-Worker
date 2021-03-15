package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
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
	upload(newPath, job)
	Busy = false

}

func upload(path string, job jobData) {
	log.Printf("Uploading %v", path)

	extraParams := map[string]string{
		"benutzer":       Cfg.LrzUser,
		"mailadresse":      Cfg.LrzMail,
		"telefon": Cfg.LrzPhone,
		"unidir": "tum",
		"subdir": Cfg.LrzSubDir,
		"info": "testupload",
	}
	request, err := newfileUploadRequest(Cfg.LrzUploadURL, extraParams, "filename", path)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	fmt.Println(request.PostForm.Encode())
	if err != nil {
		log.Fatal(err)
	} else {
		var bodyContent []byte
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		resp.Body.Read(bodyContent)
		resp.Body.Close()
		fmt.Println(string(bodyContent))
	}
	log.Println("Uploaded file to lrz.")
	pathparts := strings.Split(path, "/")
	createVodData := putVodData{
		Name:     job.Name,
		Start:    job.StreamStart,
		HlsUrl:   "https://stream.lrz.de/vod/_definst_/mp4:tum/RBG/1." + pathparts[len(pathparts)-1] + "/playlist.m3u8",
		StreamId: job.StreamId,
	}
	send, _ := json.Marshal(createVodData)
	_, err = http.Post("http://backend:7002/api/worker/putVOD/"+Cfg.WorkerID,
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

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	_ = file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fi.Name())
	if err != nil {
		return nil, err
	}
	_, _ = part.Write(fileContents)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return http.NewRequest("POST", uri, body)
}

type jobData struct {
	Id          uint      `json:"id"`
	Name        string    `json:"name"`
	StreamId    uint      `json:"streamId"`
	StreamStart time.Time `json:"streamStart"`
	Path        string    `json:"path"`
}

type putVodData struct {
	Name     string
	Start    time.Time
	HlsUrl   string
	StreamId uint
}
