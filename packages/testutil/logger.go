package testutil

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import (
	"testing"

	"github.com/iotaledger/hive.go/logger"
	"go.uber.org/zap"
)

// NewLogger produces a logger adjusted for test cases.
func NewLogger(t *testing.T) *logger.Logger {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return log.Named(t.Name()).Sugar()
}

// WithLevel returns a logger with a level increased.
// Can be useful in tests to disable logging in some parts of the system.
func WithLevel(log *logger.Logger, level logger.Level) *logger.Logger {
	return log.Desugar().WithOptions(zap.IncreaseLevel(level)).Sugar()
}
