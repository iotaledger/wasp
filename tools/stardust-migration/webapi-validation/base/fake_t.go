package base

import "fmt"

type MockT struct{}

func (m MockT) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (m MockT) FailNow() {
	panic("Test failed")
}

var T *MockT = &MockT{}
