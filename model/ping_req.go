package model

type PingReq struct {
	Workload int   `json:"workload,omitempty"`
	Jobs     []Job `json:"jobs"`
}
