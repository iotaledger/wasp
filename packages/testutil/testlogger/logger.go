// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testlogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/iotaledger/hive.go/logger"
)

type TestingT interface { // Interface so there's no need to pass the concrete type
	Name() string
}

// NewLogger produces a logger adjusted for test cases.
func NewSimple(debug bool) *logger.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("04:05.000")
	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	lvl := zapcore.InfoLevel
	if debug {
		lvl = zapcore.DebugLevel
	}
	log = log.WithOptions(zap.IncreaseLevel(lvl), zap.AddStacktrace(zapcore.FatalLevel))
	return log.Sugar()
}

// NewLogger produces a logger adjusted for test cases.
func NewLogger(t TestingT, timeLayout ...string) *logger.Logger {
	return NewNamedLogger(t.Name(), timeLayout...)
}

// NewNamedLogger produces a logger adjusted for test cases.
func NewNamedLogger(name string, timeLayout ...string) *logger.Logger {
	// log, err := zap.NewDevelopment()
	cfg := zap.NewDevelopmentConfig()
	if len(timeLayout) > 0 {
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(timeLayout[0])
	}
	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return log.Named(name).Sugar()
}

func NewSilentLogger(name string, printStackTrace bool, timeLayout ...string) *logger.Logger {
	log := NewNamedLogger(name)
	return WithLevel(log, zapcore.ErrorLevel, printStackTrace)
}

// WithLevel returns a logger with a level increased.
// Can be useful in tests to disable logging in some parts of the system.
func WithLevel(log *logger.Logger, level logger.Level, printStackTrace bool) *logger.Logger {
	if printStackTrace {
		return log.Desugar().WithOptions(zap.IncreaseLevel(level), zap.AddStacktrace(zapcore.PanicLevel)).Sugar()
	}
	return log.Desugar().WithOptions(zap.IncreaseLevel(level), zap.AddStacktrace(zapcore.FatalLevel)).Sugar()
}
