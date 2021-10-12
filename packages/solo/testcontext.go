// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import "fmt"

// TestContext is a subset of the interface provided by *testing.T and require.TestingT
// It allows to use Solo outside of unit tests.
type TestContext interface {
	Name() string
	Errorf(format string, args ...interface{})
	FailNow()
	Logf(format string, args ...interface{})
}

func NewTestContext(name string) TestContext {
	return &testContext{name}
}

type testContext struct {
	name string
}

func (f *testContext) Name() string {
	return f.name
}

func (f *testContext) Errorf(format string, args ...interface{}) {
	fmt.Printf("["+f.name+"] ERROR: "+format+"\n", args...)
}

func (f *testContext) FailNow() {
	panic("FailNow() called")
}

func (f *testContext) Logf(format string, args ...interface{}) {
	fmt.Printf("["+f.name+"] "+format+"\n", args...)
}
