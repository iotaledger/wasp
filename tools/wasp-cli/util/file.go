package util

import (
	"os"
)

func ReadFile(fname string) ([]byte, error) {
	b, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	return b, nil
}
