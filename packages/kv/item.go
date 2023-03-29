package kv

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
)

type Item struct {
	Key   Key
	Value []byte
}

type Items []Item

func (items Items) Len() int {
	return len(items)
}

func (items Items) Less(i, j int) bool {
	return items[i].Key < items[j].Key
}

func (items Items) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (it *Item) Format(format string) string {
	return fmt.Sprintf(format, iotago.EncodeHex([]byte(it.Key)), iotago.EncodeHex(it.Value))
}
