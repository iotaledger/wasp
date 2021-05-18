package dashboard

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
)

func args(args ...interface{}) []interface{} {
	return args
}

func hashref(hash hashing.HashValue) *hashing.HashValue {
	return &hash
}

func trim(max int, s string) string {
	if len(s) > max {
		s = s[:max] + "…"
	}
	// escape unprintable chars
	s = fmt.Sprintf("%q", s)
	// remove quotes
	return s[1 : len(s)-1]
}

func bytesToString(b []byte) string {
	return string(b)
}

func formatTimestamp(ts interface{}) string {
	t, ok := ts.(time.Time)
	if !ok {
		t = time.Unix(0, ts.(int64))
	}
	return t.UTC().Format(time.RFC3339)
}

func exploreAddressUrl(baseUrl string) func(address ledgerstate.Address) string {
	return func(address ledgerstate.Address) string {
		return baseUrl + "/" + address.Base58()
	}
}
