// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testpeers

import (
	"embed"
	"fmt"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

//go:embed testkeys_pregenerated-*.bin
var embedded embed.FS

func pregeneratedDksName(n, t uint16) string {
	return fmt.Sprintf("testkeys_pregenerated-%v-%v.bin", n, t)
}

func pregeneratedDksRead(n, t uint16) [][]byte {
	var err error
	var buf []byte
	if buf, err = embedded.ReadFile(pregeneratedDksName(n, t)); err != nil {
		panic(err)
	}
	rr := rwutil.NewBytesReader(buf)
	bufN := rr.ReadSize16()
	if rr.Err != nil {
		panic(rr.Err)
	}
	if int(n) != bufN {
		panic("wrong_file")
	}
	res := make([][]byte, n)
	for i := range res {
		res[i] = rr.ReadBytes()
		if rr.Err != nil {
			panic(rr.Err)
		}
	}
	return res
}
