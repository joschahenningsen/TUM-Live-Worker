package silencedetect

import (
	"TUM-Live-Worker/model"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SilenceDetect struct {
	Input    string
	Silences *[]model.silence
}

func New(input string) *SilenceDetect {
	return &SilenceDetect{Input: input}
}

func (s *SilenceDetect) ParseSilence() error {
	log.Println("Start detecting silence in eist2021_05_20_08_00COMB")
	cmd := exec.Command("ffmpeg", "-nostats", "-i", s.Input, "-af", "silencedetect=n=-30dB:d=10", "-f", "null", "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%v", err)
		return err
	}
	l := strings.Split(string(output), "\n")
	var silences []model.silence
	startRe, _ := regexp.Compile("silence_start: [0-9]+(\\.[0-9]+)?")
	endRe, _ := regexp.Compile("silence_end: [0-9]+(\\.[0-9]+)? \\| silence_duration: [0-9]+(\\.[0-9]+)?")
	for _, str := range l {
		if startRe.MatchString(str) {
			start, err := strconv.ParseFloat(strings.Split(str, "silence_start: ")[1], 32)
			if err != nil {
				log.Printf("%v", err)
				return err
			}
			silences = append(silences, model.silence{
				Start: uint(start),
				End:   0,
			})
		} else if endRe.MatchString(str) {
			end, err := strconv.ParseFloat(strings.Split(strings.Split(str, "silence_end: ")[1], " |")[0], 32)
			if err != nil || silences == nil || len(silences) == 0 {
				log.Printf("%v", err)
				return err
			}
			silences[len(silences)-1].End = uint(end)
			silences[len(silences)-1].Len = time.Duration(uint(end)-silences[len(silences)-1].Start) * time.Millisecond * 1000
		}
	}
	for _, curSilence := range silences {
		log.Printf("[found silence]: %v->%v Duration: %v", curSilence.Start, curSilence.End, curSilence.Len)
	}
	s.Silences = &silences
	s.postprocess()
	return nil
}

func (s *SilenceDetect) postprocess() {
	if len(*s.Silences) < 2 {
		return
	}
}
