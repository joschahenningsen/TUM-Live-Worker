package main

import (
	"TUM-Live-Worker/model"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	Workload = 0
	Status   = "idle"
	OsSignal chan os.Signal
	Cfg      Config
	Jobs     []model.Job
)

func main() {
	log.Println("Starting worker")
	Jobs = []model.Job{}
	OsSignal = make(chan os.Signal, 1)
	Cfg = Config{
		LrzUser:      os.Getenv("LRZ_USER"),
		LrzMail:      os.Getenv("LRZ_MAIL"),
		LrzPhone:     os.Getenv("LRZ_PHONE"),
		LrzSubDir:    os.Getenv("LRZ_SUBDIR"),
		LrzUploadURL: os.Getenv("LRZ_UPLOAD_URL"),
		WorkerID:     os.Getenv("WORKERID"),
		IngestBase:   os.Getenv("INGEST_BASE"),
		MainBase:     os.Getenv("MAIN_BASE"),
	}
	cronService := cron.New()
	_, _ = cronService.AddFunc("0-59/5 * * * *", ping)
	cronService.Start()
	configRouter()
	LoopForever()
}

// LoopForever on signal processing
func LoopForever() {
	fmt.Printf("Entering infinite loop\n")

	signal.Notify(OsSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	_ = <-OsSignal

	fmt.Printf("Exiting infinite loop received OsSignal\n")
}

type Config struct {
	LrzUser      string
	LrzMail      string
	LrzPhone     string
	LrzSubDir    string
	LrzUploadURL string
	WorkerID     string
	IngestBase   string
	MainBase     string
	Cert         string
	Key          string
}
