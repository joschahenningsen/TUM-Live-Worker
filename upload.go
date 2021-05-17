package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func upload(path string, streamID string, version string) {
	pathparts := strings.Split(path, "/")
	values := map[string]io.Reader{
		"filename":    mustOpen(path),
		"benutzer":    strings.NewReader(Cfg.LrzUser),
		"mailadresse": strings.NewReader(Cfg.LrzMail),
		"telefon":     strings.NewReader(Cfg.LrzPhone),
		"unidir":      strings.NewReader("tum"),
		"subdir":      strings.NewReader(Cfg.LrzSubDir),
		"info":        strings.NewReader(""),
	}
	err := PostFileUpload(Cfg.LrzUploadURL, values)
	if err != nil {
		log.Printf("%s", err)
		return
	}
	createVodData := putVodData{
		HlsUrl:   "https://stream.lrz.de/vod/_definst_/mp4:tum/RBG/" + pathparts[len(pathparts)-1] + "/playlist.m3u8",
		FilePath: path,
		StreamId: streamID,
		Version:  version,
	}
	send, _ := json.Marshal(createVodData)
	_, err = http.Post(fmt.Sprintf("%s/api/worker/putVOD/%s", Cfg.MainBase, Cfg.WorkerID),
		"application/json",
		bytes.NewBuffer(send))
	if err != nil {
		log.Printf("%v\n", err)
	}
}

// PostFileUpload - example kindly provided by Attila O. Thanks buddy!
func PostFileUpload(url string, values map[string]io.Reader) (err error) {
	client := http.DefaultClient
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

type putVodData struct {
	HlsUrl   string
	Version  string
	FilePath string
	StreamId string
}
