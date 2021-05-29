package silencedetect

import (
	"TUM-Live-Worker/model"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type SilenceDetect struct {
	Input    string
	Silences *[]model.Silence
}

func New(input string) *SilenceDetect {
	return &SilenceDetect{Input: input}
}

func (s *SilenceDetect) ParseSilence() error {
	log.Printf("Start detecting silence in %s", s.Input)
	cmd := exec.Command("nice", "ffmpeg", "-nostats", "-i", s.Input, "-af", "silencedetect=n=-15dB:d=30", "-f", "null", "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	log.Println(string(output))
	l := strings.Split(string(output), "\n")
	var silences []model.Silence
	for _, str := range l {
		if strings.Contains(str, "silence_start:") {
			start, err := strconv.ParseFloat(strings.Split(str, "silence_start: ")[1], 32)
			if err != nil {
				return err
			}
			silences = append(silences, model.Silence{
				Start: uint(start),
				End:   0,
			})
		} else if strings.Contains(str, "silence_end:") {
			end, err := strconv.ParseFloat(strings.Split(strings.Split(str, "silence_end: ")[1], " |")[0], 32)
			if err != nil || silences == nil || len(silences) == 0 {
				return err
			}
			silences[len(silences)-1].End = uint(end)
		}
	}

	s.Silences = &silences
	s.postprocess()
	slc := *s.Silences
	for _, silence := range slc {
		log.Printf("%v, %v", silence.Start, silence.End)
	}
	return nil
}

//postprocess merges short duration of sound into units of silence
func (s *SilenceDetect) postprocess() {
	oldSilences := *s.Silences
	if len(oldSilences) < 2 {
		return
	}
	if oldSilences[0].Start < 30 {
		oldSilences[0].Start = 0
	}
	newSilences := []model.Silence{{Start: oldSilences[0].Start, End: oldSilences[0].Start}}
	oldPtr := 0
	for oldPtr < len(oldSilences) {
		if oldSilences[oldPtr].Start-newSilences[len(newSilences)-1].End < 30 { // Ignore sound that's shorter than 30 seconds
			newSilences[len(newSilences)-1].End = oldSilences[oldPtr].End
		} else {
			newSilences = append(newSilences, oldSilences[oldPtr])
		}
		oldPtr++
	}
	s.Silences = &newSilences
}
