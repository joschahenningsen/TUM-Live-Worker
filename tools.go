package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

func ping() {
	_, err := http.Post(fmt.Sprintf("%v/api/worker/ping/%v", Cfg.MainBase, Cfg.WorkerID), "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Println("Couldn't ping main")
		return
	}
}
