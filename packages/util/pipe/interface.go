package pipe

// minQueueLen is smallest capacity that queue may have.
const minQueueLen = 16

type Hashable interface {
	// For requirements of this function see https://docs.oracle.com/javase/8/docs/api/java/lang/Object.html#hashCode--
	// Additional requirement: the returned value must be valid as a map key
	GetHash() interface{}
	// For requirements of this function see https://docs.oracle.com/javase/8/docs/api/java/lang/Object.html#equals-java.lang.Object-
	Equals(elem interface{}) bool
}

type Set interface {
	Size() int
	Add(elem interface{}) bool
	Clear()
	Contains(elem interface{}) bool
	Remove(elem interface{}) bool
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
