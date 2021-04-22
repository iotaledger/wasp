package kv

type Item struct {
	Key   Key
	Value []byte
}

type Items []Item

func (items Items) Len() int {
	return len(items)
}

func (items Items) Less(i int, j int) bool {
	return items[i].Key < items[j].Key
}

func (items Items) Swap(i int, j int) {
	items[i], items[j] = items[j], items[i]
}
