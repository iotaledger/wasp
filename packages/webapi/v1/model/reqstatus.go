package model

import (
	"time"
)

type WaitRequestProcessedParams struct {
	Timeout time.Duration `swagger:"desc(Timeout in nanoseconds),default(30 seconds)"`
}

type RequestReceiptResponse struct {
	Receipt string `swagger:"desc(Request receipt, empty if request was not processed)"`
}

const WaitRequestProcessedDefaultTimeout = 30 * time.Second
