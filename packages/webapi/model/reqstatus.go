package model

import "time"

type WaitRequestProcessedParams struct {
	Timeout time.Duration `swagger:"desc(Timeout in nanoseconds),default(30 seconds)"`
}

type RequestStatusResponse struct {
	IsProcessed bool `swagger:"desc(True if the request has been processed)"`
}

const WaitRequestProcessedDefaultTimeout = 30 * time.Second
