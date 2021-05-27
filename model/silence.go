package model

import "time"

type Silence struct {
	Start uint          `json:"start,omitempty"`
	End   uint          `json:"end,omitempty"`
	Len   time.Duration `json:"len,omitempty"`
}
