// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"time"

	"github.com/iotaledger/wasp/packages/gpa"
)

type inputTimeData struct {
	timeData time.Time
}

func NewInputTimeData(timeData time.Time) gpa.Input {
	return &inputTimeData{timeData: timeData}
}
