// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testpeers

import (
	"bytes"
	"embed"
	"fmt"

	"github.com/iotaledger/wasp/packages/util"
)

//go:embed testkeys_pregenerated-*.bin
var embedded embed.FS

func pregeneratedDksName(n, t uint16) string {
	return fmt.Sprintf("testkeys_pregenerated-%v-%v.bin", n, t)
}

func pregeneratedDksRead(n, t uint16) [][]byte {
	var err error
	var buf []byte
	var bufN uint16
	if buf, err = embedded.ReadFile(pregeneratedDksName(n, t)); err != nil {
		panic(err)
	}
	r := bytes.NewReader(buf)
	if err = util.ReadUint16(r, &bufN); err != nil {
		panic(err)
	}
	if n != bufN {
		panic("wrong_file")
	}
	res := make([][]byte, n)
	for i := range res {
		if res[i], err = util.ReadBytes16(r); err != nil {
			panic(r)
		}
	}
	return res
}
