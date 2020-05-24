package sctransaction

import (
	"bytes"
	"errors"
	"github.com/iotaledger/wasp/packages/hashing"
	"io"
)

func ReadRequestId(r io.Reader, reqid *RequestId) error {
	n, err := r.Read(reqid[:])
	if err != nil {
		return err
	}
	if n != RequestIdSize {
		return errors.New("error while reading request id")
	}
	return nil
}

func BatchHash(reqids []RequestId) hashing.HashValue {
	var buf bytes.Buffer
	for i := range reqids {
		buf.Write(reqids[i][:])
	}
	return *hashing.HashData(buf.Bytes())
}
