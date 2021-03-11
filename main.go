package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"os"
	"os/signal"
	"syscall"
)

var (
	Busy bool = false
	OsSignal chan os.Signal
	WorkerID string
)

func main() {
	OsSignal = make(chan os.Signal, 1)
	WorkerID = os.Getenv("WORKERID")
	cronService := cron.New()
	_, _ = cronService.AddFunc("* * * * *", convertAndUpload)
	cronService.Start()
	LoopForever()
}

// LoopForever on signal processing
func LoopForever() {
	fmt.Printf("Entering infinite loop\n")

	signal.Notify(OsSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	_ = <-OsSignal

	fmt.Printf("Exiting infinite loop received OsSignal\n")
}
