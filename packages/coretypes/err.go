package coretypes

import "errors"

var (
	ErrWrongDataConversion = errors.New("wrong data conversion")
	ErrWrongDataLength     = errors.New("wrong data length")
)
