package model

type PingReq struct {
	Workload int `json:"workload"`
	Status   string `json:"status"`
}
