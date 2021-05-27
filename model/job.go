package model

type Job struct {
	Title    string  `json:"title,omitempty"`
	PID      *string `json:"pid,omitempty"`
	Progress int     `json:"progress,omitempty"`
}
