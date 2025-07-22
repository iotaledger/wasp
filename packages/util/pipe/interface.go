package pipe

import "github.com/iotaledger/wasp/v2/packages/hashing"

type Hashable interface {
	GetHash() hashing.HashValue
}

type Queue[E any] interface {
	Length() int
	Add(elem E) bool
	Peek() E
	Get(i int) E
	Remove() E
}

type Deque[E any] interface {
	Length() int
	AddStart(elem E) bool
	AddEnd(elem E) bool
	PeekStart() E
	PeekEnd() E
	PeekNStart(n int) []E
	PeekNEnd(n int) []E
	PeekAll() []E
	Get(i int) E
	RemoveStart() E
	RemoveEnd() E
	RemoveAt(i int) E
}

type Pipe[E any] interface {
	In() chan<- E
	Out() <-chan E
	Len() int
	Close()
	Discard()
	TryAdd(e E, log func(msg string, args ...interface{}))
}
