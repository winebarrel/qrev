package qrev

type Status string

const (
	StatusDone Status = "done"
	StatusFail Status = "fail"
	StatusSkip Status = "skip"
)
