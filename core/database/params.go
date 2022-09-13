package database

import (
	"github.com/iotaledger/hive.go/core/app"
)

type ParametersDatabase struct {
	InMemory  bool   `default:"false" usage:"whether the database is only kept in memory and not persisted"`
	Directory string `default:"waspdb" usage:"path to the database folder"`
}

var ParamsDatabase = &ParametersDatabase{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"database": ParamsDatabase,
	},
	Masked: nil,
}
