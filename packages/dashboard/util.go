package dashboard

import (
	"fmt"
	"strings"
	"time"
)

func FormatTimestamp(ts interface{}) string {
	t, ok := ts.(time.Time)
	if !ok {
		t = time.Unix(0, ts.(int64))
	}
	return t.UTC().Format(time.RFC3339)
}

func ExploreAddressUrl(baseUrl string) func(address fmt.Stringer) string {
	return func(address fmt.Stringer) string {
		return baseUrl + "/" + address.String()
	}
}

func ExploreAddressUrlFromGoshimmerUri(uri string) string {
	url := strings.Split(uri, ":")[0] + ":8081/explorer/address"
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
}

const TplExploreAddress = `{{define "address"}}<code>{{.}}</code> <a href="{{exploreAddressUrl .}}">[+]</a>{{end}}`
