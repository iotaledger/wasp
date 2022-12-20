package pipe

import "github.com/iotaledger/wasp/packages/hashing"

// LimitedPriorityHashQueue is a queue, which can prioritize elements,
// limit its growth and reject already included elements.
type LimitedPriorityHashQueue[E Hashable] struct {
	buf         []E
	head        int
	pend        int
	tail        int
	count       int
	priorityFun func(E) bool
	limit       int
	hashMap     *map[hashing.HashValue]bool
}

var _ Queue[Hashable] = &LimitedPriorityHashQueue[Hashable]{}

const Infinity = 0

func NewDefaultLimitedPriorityHashQueue[E Hashable]() Queue[E] {
	return NewHashLimitedPriorityHashQueue[E](false)
}

func NewPriorityLimitedPriorityHashQueue[E Hashable](priorityFun func(E) bool) Queue[E] {
	return NewPriorityHashLimitedPriorityHashQueue(priorityFun, false)
}

func NewLimitLimitedPriorityHashQueue[E Hashable](limit int) Queue[E] {
	return NewLimitHashLimitedPriorityHashQueue[E](limit, false)
}

func NewLimitPriorityLimitedPriorityHashQueue[E Hashable](priorityFun func(E) bool, limit int) Queue[E] {
	return NewLimitedPriorityHashQueue(priorityFun, limit, false)
}

func NewHashLimitedPriorityHashQueue[E Hashable](hashNeeded bool) Queue[E] {
	return NewLimitHashLimitedPriorityHashQueue[E](Infinity, hashNeeded)
}

func NewPriorityHashLimitedPriorityHashQueue[E Hashable](priorityFun func(E) bool, hashNeeded bool) Queue[E] {
	return NewLimitedPriorityHashQueue(priorityFun, Infinity, hashNeeded)
}

func NewLimitHashLimitedPriorityHashQueue[E Hashable](limit int, hashNeeded bool) Queue[E] {
	return NewLimitedPriorityHashQueue(func(E) bool { return false }, limit, hashNeeded)
}

func NewLimitedPriorityHashQueue[E Hashable](priorityFun func(E) bool, limit int, hashNeeded bool) Queue[E] {
	var initBufSize int
	if (limit != Infinity) && (limit < minQueueLen) {
		initBufSize = limit
	} else {
		initBufSize = minQueueLen
	}
	var hashMap *map[hashing.HashValue]bool
	if hashNeeded {
		hMap := make(map[hashing.HashValue]bool)
		hashMap = &hMap
	} else {
		hashMap = nil
	}
	return &LimitedPriorityHashQueue[E]{
		head:        0,
		pend:        -1,
		tail:        0,
		count:       0,
		buf:         make([]E, initBufSize),
		priorityFun: priorityFun,
		limit:       limit,
		hashMap:     hashMap,
	}
}

// Length returns the number of elements currently stored in the queue.
func (q *LimitedPriorityHashQueue[E]) Length() int {
	return q.count
}

func (q *LimitedPriorityHashQueue[E]) getIndex(rawIndex int) int {
	index := rawIndex % len(q.buf)
	if index < 0 {
		return index + len(q.buf)
	}
	return index
}

// resizes the queue to fit exactly twice its current contents
// this can result in shrinking if the queue is less than half-full
// the size of the resized queue is never smaller than minQueueLen, except
// when the limit is smaller.
func (q *LimitedPriorityHashQueue[E]) resize() {
	newSize := q.count << 1
	if newSize < minQueueLen {
		newSize = minQueueLen
	}
	if (q.limit != Infinity) && (newSize > q.limit) {
		newSize = q.limit
	}
	newBuf := make([]E, newSize)

	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}

	if q.pend >= 0 {
		q.pend = q.getIndex(q.pend - q.head)
	}
	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

// Add puts an element to the start or end of the queue, depending
// on the result of priorityFun. If the limited queue is full it adds a new element
// removing the previously added element, according to the following rules:
//   - not prioritized element is chosen for deletion, if possible
//   - the chosen for deletion element is always the oldest among its type
//   - not prioritized element can not be added if there are no not prioritized
//     elements to delete
//
// If it is a hash queue, the element is not added, if it is already in the queue.
// If the add was successful, returns `true`.
//

//nolint:gocyclo
func (q *LimitedPriorityHashQueue[E]) Add(elem E) bool {
	var elemHash hashing.HashValue
	if q.hashMap != nil {
		elemHash = elem.GetHash()
		contains, ok := (*q.hashMap)[elemHash]
		if ok && contains {
			// duplicate element; ignoring
			return false
		}
	}
	limitReached := false
	if q.count == len(q.buf) {
		if (q.limit != Infinity) && (q.count >= q.limit) {
			limitReached = true
		} else {
			q.resize()
		}
	}
	priority := q.priorityFun(elem)
	if limitReached && !priority && (q.pend == q.getIndex(q.tail-1)) {
		// Not possible to add not priority element in queue full of priority elements
		return false
	}
	if limitReached {
		var deleteElem interface{}
		if q.pend < 0 {
			deleteElem = q.buf[q.head]
			q.head = q.getIndex(q.head + 1)
		} else {
			ptail := q.getIndex(q.pend + 1)
			if ptail == q.tail {
				deleteElem = q.buf[q.pend]
				q.tail = q.getIndex(q.tail - 1)
				q.pend = q.getIndex(q.pend - 1)
			} else {
				deleteElem = q.buf[ptail]
				if ptail > q.head {
					copy(q.buf[q.head+1:ptail+1], q.buf[q.head:ptail])
				} else {
					oldHead := q.buf[q.head]
					if ptail > 0 {
						copy(q.buf[1:ptail+1], q.buf[:ptail])
					}
					lastIndex := len(q.buf) - 1
					q.buf[0] = q.buf[lastIndex]
					if q.head < lastIndex {
						copy(q.buf[q.head+1:], q.buf[q.head:lastIndex])
					}
					q.buf[q.getIndex(q.head+1)] = oldHead
				}
				q.pend = q.getIndex(q.pend + 1)
				q.head = q.getIndex(q.head + 1)
			}
		}
		if q.hashMap != nil {
			deleteElemHashable, ok := deleteElem.(Hashable)
			if !ok {
				panic("Deleting not hashable element")
			}
			delete(*q.hashMap, deleteElemHashable.GetHash())
		}
	}
	if priority {
		q.head = q.getIndex(q.head - 1)
		q.buf[q.head] = elem
		if q.pend < 0 {
			q.pend = q.head
		}
	} else {
		q.buf[q.tail] = elem
		// bitwise modulus
		q.tail = q.getIndex(q.tail + 1)
	}
	if !limitReached {
		q.count++
	}
	if q.hashMap != nil {
		(*q.hashMap)[elemHash] = true
	}
	return true
}

// Peek returns the element at the head of the queue. This call panics
// if the queue is empty.
func (q *LimitedPriorityHashQueue[E]) Peek() E {
	if q.count <= 0 {
		panic("queue: Peek() called on empty queue")
	}
	return q.buf[q.head]
}

// Get returns the element at index i in the queue. If the index is
// invalid, the call will panic. This method accepts both positive and
// negative index values. Index 0 refers to the first element, and
// index -1 refers to the last.
func (q *LimitedPriorityHashQueue[E]) Get(i int) E {
	// If indexing backwards, convert to positive index.
	if i < 0 {
		i += q.count
	}
	if i < 0 || i >= q.count {
		panic("queue: Get() called with index out of range")
	}
	// bitwise modulus
	return q.buf[q.getIndex(q.head+i)]
}

// Remove removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *LimitedPriorityHashQueue[E]) Remove() E {
	if q.count <= 0 {
		panic("queue: Remove() called on empty queue")
	}
	ret := q.buf[q.head]
	var nilE E
	q.buf[q.head] = nilE
	if q.head == q.pend {
		q.pend = -1
	}
	q.head = q.getIndex(q.head + 1)
	q.count--
	// Resize down if buffer 1/4 full.
	if (len(q.buf) > minQueueLen) && ((q.count << 2) <= len(q.buf)) {
		q.resize()
	}
	if q.hashMap != nil {
		delete(*q.hashMap, ret.GetHash())
	}
	return ret
}
