package logger

import (
	"github.com/iotaledger/hive.go/core/app"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name: "Logger",
			//Params:    params,
			Configure: configure,
		},
	}
}

var (
	CoreComponent *app.CoreComponent
)

func configure() error {
	initGoEthLogger(CoreComponent.Logger().Named("go-ethereum"))

	return nil
}
