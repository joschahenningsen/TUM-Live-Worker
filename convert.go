package main

import (
	"log"
	"os/exec"
)

func convert(inFile string, outFile string) {
	Workload++
	Status = "Converting " + inFile
	ping()
	defer func() {
		Workload--
		if Workload == 0 {
			Status = "idle"
		}
		ping()
	}()
	log.Printf("Processing %v output: %v", inFile, outFile)
	cmd := exec.Command("nice", "ffmpeg", "-i", inFile, "-preset", "veryfast","-c:v", "libx264", "-c:a", "aac", "-ar", "44100", outFile)
	log.Printf("%s", cmd.String())

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
	// _ = os.Remove(inFile)
}