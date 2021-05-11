package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func ping() {
	req := pingReq{Status: Status, Workload: Workload}
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

type pingReq struct {
	Workload int    `json:"workload,omitempty"`
	Status   string `json:"status,omitempty"`
}
