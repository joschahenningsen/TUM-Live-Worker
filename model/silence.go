package model

type Silence struct {
	Start uint `json:"start"`
	End   uint `json:"end"`
}

type SilenceReq struct {
	StreamID string    `json:"stream_id"`
	Silences []Silence `json:"silences"`
}
