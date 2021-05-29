package model

type PingReq struct {
	Workload int `json:"workload,omitempty"`
	Status   string `json:"status"`
}
