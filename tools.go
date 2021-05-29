package main

import (
	"TUM-Live-Worker/model"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func ping() {
	req := model.PingReq{Workload: Workload, Jobs: Jobs}
	marshal, err := json.Marshal(&req)
	if err != nil {
		log.Printf("couldn't marshal ping request")
		return
	}
	_, err = http.Post(fmt.Sprintf("%v/api/worker/ping/%v", Cfg.MainBase, Cfg.WorkerID), "application/json", bytes.NewBuffer(marshal))
	if err != nil {
		log.Println("Couldn't ping main")
		return
	}
}

func notifySilenceDetectionResults(silences *[]model.Silence, streamID string) {
	reqObj := model.SilenceReq{Silences: *silences, StreamID: streamID}
	req, err := json.Marshal(reqObj)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	log.Println(string(req))
	_, err = http.Post(fmt.Sprintf("%v/api/worker/silenceResults/%v", Cfg.MainBase, Cfg.WorkerID), "application/json", bytes.NewBuffer(req))
	if err != nil {
		log.Printf("%v", err)
	}
}
