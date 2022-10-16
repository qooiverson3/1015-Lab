package domain

import "time"

type RequestHeader struct {
	Token string `header:"token" binding:"required"`
}

type RequestBodyViaSequence struct {
	Stream   string `json:"stream" binding:"required" example:"TestStream"`
	StartSeq uint64 `json:"startSeq" example:"10000" binding:"required"`
	EndSeq   uint64 `json:"endSeq" example:"12000"`
	Phase    uint   `json:"phase" example:"0" binding:"required"`
}

type RequestBodyViaTimeStamp struct {
	Stream string `json:"stream" binding:"required" example:"TestStream"`
	StartT uint64 `json:"starT" example:"1664520600" binding:"required"`
	EndT   uint64 `json:"endT" example:"1664520600"`
	Phase  uint   `json:"phase" example:"0" binding:"required"`
}

type ResponseOfGetRestoreSuccess struct {
	Status bool      `json:"status"`
	Data   []Restore `json:"list"`
}

type ResponseOfRestoreForbidden struct {
	Status bool      `json:"status"`
	Data   []Restore `json:"list"`
}

type ResponseOfPostRestoreSuccess struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type Filter struct {
	Seq          FilterBySeq
	Timeinterval FilterByTimeInterval
}

type FilterBySeq struct {
	Start uint64 `json:"start"`
	End   uint64 `json:"end"`
}

type FilterByTimeInterval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
