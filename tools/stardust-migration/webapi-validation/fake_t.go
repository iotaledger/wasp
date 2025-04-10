package webapi_validation

import "fmt"

type mockT struct{}

func (m mockT) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (m mockT) FailNow() {
	panic("Test failed")
}

var t *mockT = &mockT{}
