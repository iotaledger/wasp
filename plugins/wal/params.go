package wal

import (
	"github.com/iotaledger/hive.go/core/app"
)

type ParametersWAL struct {
	Enabled   bool   `default:"true" usage:"whether the WAL plugin is enabled"`
	Directory string `default:"wal" usage:"path to logs folder"`
}

var ParamsWAL = &ParametersWAL{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"wal": ParamsWAL,
	},
	Masked: nil,
}
