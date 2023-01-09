// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package dkg

// InvalidParamsError is used to distinguish user errors from the execution errors.
type InvalidParamsError struct {
	error
}

func (e InvalidParamsError) Error() string {
	return e.error.Error()
}

func invalidParams(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(InvalidParamsError); ok {
		return e
	}
	return InvalidParamsError{err}
}
