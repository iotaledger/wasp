package kv

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
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
	return fmt.Sprintf(format, hexutil.Encode([]byte(it.Key)), hexutil.Encode(it.Value))
}
