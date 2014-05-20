package data

import (
	"time"
)

type ServiceResults struct {
	CheckName string `json:"name"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	ReturnCode uint8 `json:"returncode"`
	Time time.Time `json:"time"`
}
