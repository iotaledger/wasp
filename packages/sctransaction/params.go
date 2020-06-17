package sctransaction

import (
	"fmt"
	"io"
	"sort"

	"github.com/iotaledger/wasp/packages/util"
)

type Params map[string][]byte

func (p Params) SetInt64(k string, v int64) {
	p[k] = util.Uint64To8Bytes(uint64(v))
}

func (p Params) GetInt64(k string) (int64, bool) {
	v, ok := p[k]
	if !ok {
		return 0, false
	}
	if len(v) != 8 {
		return 0, false
	}
	return int64(util.Uint64From8Bytes(v)), true
}

func (p Params) SetString(k string, v string) {
	p[k] = []byte(v)
}

func (p Params) GetString(k string) (string, bool) {
	v, ok := p[k]
	if !ok {
		return "", false
	}
	return string(v), true
}

func (p Params) sortedKeys() []string {
	keys := make([]string, 0)
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (p Params) Write(w io.Writer) error {
	if len(p) > util.MaxUint16 {
		return fmt.Errorf("Too many parameters")
	}

	// need to produce a deterministic output
	keys := p.sortedKeys()

	err := util.WriteUint16(w, uint16(len(p)))
	if err != nil {
		return err
	}
	for _, k := range keys {
		err = util.WriteString16(w, k)
		if err != nil {
			return err
		}
		err = util.WriteBytes16(w, p[k])
		if err != nil {
			return err
		}
	}

	return nil
}

func (p Params) Read(r io.Reader) error {
	var n uint16
	err := util.ReadUint16(r, &n)
	if err != nil {
		return err
	}

	for i := uint16(0); i < n; i++ {
		k, err := util.ReadString16(r)
		if err != nil {
			return err
		}
		v, err := util.ReadBytes16(r)
		if err != nil {
			return err
		}
		p[k] = v
	}

	return nil
}
