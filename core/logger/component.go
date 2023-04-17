package logger

import (
	"github.com/iotaledger/hive.go/app"
)

func init() {
	Component = &app.Component{
		Name:      "Logger",
		Configure: configure,
	}
}

var Component *app.Component

func configure() error {
	initGoEthLogger(Component.App().NewLogger("go-ethereum"))

	return nil
}
