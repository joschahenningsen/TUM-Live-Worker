package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joschahenningsen/TUM-Live-Worker/model"
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
