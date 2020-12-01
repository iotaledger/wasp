package dkg

// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import "go.dedis.ch/kyber/v3"

func pubToBytes(pub kyber.Point) ([]byte, error) {
	return pub.MarshalBinary()
}

func pubsToBytes(pubs []kyber.Point) ([][]byte, error) {
	var bytes = make([][]byte, len(pubs))
	for i := range pubs {
		if b, err := pubToBytes(pubs[i]); err == nil {
			bytes[i] = b
		} else {
			return nil, err
		}
	}
	return bytes, nil
}

func pubFromBytes(bytes []byte, suite kyber.Group) (kyber.Point, error) {
	pubKey := suite.Point()
	if err := pubKey.UnmarshalBinary(bytes); err != nil {
		return nil, err
	}
	return pubKey, nil
}

func pubsFromBytes(bytes [][]byte, suite kyber.Group) ([]kyber.Point, error) {
	var pubs = make([]kyber.Point, len(bytes))
	for i := range pubs {
		if b, err := pubFromBytes(bytes[i], suite); err == nil {
			pubs[i] = b
		} else {
			return nil, err
		}
	}
	return pubs, nil
}

func haveAll(buf []bool) bool {
	for i := range buf {
		if !buf[i] {
			return false
		}
	}
	return true
}
