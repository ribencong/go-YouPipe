package service

import (
	"github.com/youpipe/go-youPipe/account"
	"time"
)

type License struct {
	signature string
	startDate time.Time
	endDate   time.Time
	userAddr  account.ID
}
