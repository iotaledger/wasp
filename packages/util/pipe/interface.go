package pipe

import "github.com/iotaledger/wasp/packages/hashing"

// minQueueLen is smallest capacity that queue may have.
const minQueueLen = 16

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

type Pipe[E any] interface {
	In() chan<- E
	Out() <-chan E
	Len() int
	Close()
}
