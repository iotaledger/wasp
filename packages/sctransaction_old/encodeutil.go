package sctransaction_old

import (
	"errors"
)

// the byte is needed for the parses to quickly recognize
// what kind of block it is: state or request
// max number of request blocks in the transaction is 127

const stateBlockMask = byte(0x80)

func encodeMetaByte(hasState bool, numRequests byte) (byte, error) {
	if numRequests > 127 {
		return 0, errors.New("can't be more than 127 requests")
	}
	ret := numRequests
	if hasState {
		ret = ret | stateBlockMask
	}
	return ret, nil
}

func decodeMetaByte(b byte) (bool, byte) {
	return b&stateBlockMask != 0, b & ^stateBlockMask
}
