// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package testlogger provides logging utilities for testing purposes
package testlogger

import (
	"log/slog"

	"github.com/iotaledger/hive.go/log"
)

type TestingT interface { // Interface so there's no need to pass the concrete type
	Name() string
}

// NewLogger produces a logger adjusted for test cases.
func NewLogger(t TestingT, timeLayout ...string) log.Logger {
	return NewNamedLogger(t.Name(), timeLayout...)
}

// NewNamedLogger produces a logger adjusted for test cases.
func NewNamedLogger(name string, timeLayout ...string) log.Logger {
	// log, err := zap.NewDevelopment()
	logger := log.NewLogger()
	if len(timeLayout) > 0 {
		logger = log.NewLogger(log.WithTimeFormat(timeLayout[0]))
	}

	return logger.NewChildLogger(name)
}

func NewSilentLogger(name string, printStackTrace bool, timeLayout ...string) log.Logger {
	log := NewNamedLogger(name)
	return WithLevel(log, slog.LevelError, printStackTrace)
}

// WithLevel returns a logger with a level increased.
// Can be useful in tests to disable logging in some parts of the system.
func WithLevel(log log.Logger, level log.Level, printStackTrace bool) log.Logger {
	log.SetLogLevel(level)
	return log
}
