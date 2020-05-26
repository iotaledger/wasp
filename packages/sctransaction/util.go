package sctransaction

import (
	"errors"
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
