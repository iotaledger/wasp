package pipe

import "github.com/iotaledger/wasp/packages/hashing"

// minQueueLen is smallest capacity that queue may have.
const minQueueLen = 16

type Hashable interface {
	GetHash() hashing.HashValue
}

type Queue interface {
	Length() int
	Add(elem interface{}) bool
	Peek() interface{}
	Get(i int) interface{}
	Remove() interface{}
}

type Pipe interface {
	In() chan<- interface{}
	Out() <-chan interface{}
	Len() int
	Close()
}
