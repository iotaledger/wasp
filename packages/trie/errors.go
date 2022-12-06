package trie

import "errors"

var (
	ErrWrongNibble         = errors.New("key16 byte must be less than 0x0F")
	ErrEmpty               = errors.New("encoded key16 can't be empty")
	ErrWrongFormat         = errors.New("encoded key16 wrong format")
	ErrNotAllBytesConsumed = errors.New("serialization error: not all bytes were consumed")
)
