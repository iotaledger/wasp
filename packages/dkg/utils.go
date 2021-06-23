// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg

import "go.dedis.ch/kyber/v3"

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

//nolint:unused
func pubToBytes(pub kyber.Point) ([]byte, error) {
	return pub.MarshalBinary()
}

//nolint:unused,deadcode
func pubsToBytes(pubs []kyber.Point) ([][]byte, error) {
	bytes := make([][]byte, len(pubs))
	for i := range pubs {
		if b, err := pubToBytes(pubs[i]); err == nil {
			bytes[i] = b
		} else {
			return nil, err
		}
	}
	return bytes, nil
}

//nolint:unused
func pubFromBytes(bytes []byte, suite kyber.Group) (kyber.Point, error) {
	pubKey := suite.Point()
	if err := pubKey.UnmarshalBinary(bytes); err != nil {
		return nil, err
	}
	return pubKey, nil
}

//nolint:unused,deadcode
func pubsFromBytes(bytes [][]byte, suite kyber.Group) ([]kyber.Point, error) {
	pubs := make([]kyber.Point, len(bytes))
	for i := range pubs {
		if b, err := pubFromBytes(bytes[i], suite); err == nil {
			pubs[i] = b
		} else {
			return nil, err
		}
	}
	return pubs, nil
}

//nolint:unused,deadcode
func haveAll(buf []bool) bool {
	for i := range buf {
		if !buf[i] {
			return false
		}
	}
	return true
}
