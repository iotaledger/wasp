package pipe

import (
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

// LimitedPriorityHashQueue is a queue, which can prioritize elements,
// limit its growth and reject already included elements.
type LimitedPriorityHashQueue[E any] struct {
	deque       Deque[E]
	pend        int // points to the next element after last priority element
	priorityFun func(E) bool
	hashMap     *shrinkingmap.ShrinkingMap[hashing.HashValue, struct{}]
}

var _ Queue[Hashable] = &LimitedPriorityHashQueue[Hashable]{}

func NewLimitedPriorityHashQueue[E any]() Queue[E] {
	return NewLimitLimitedPriorityHashQueue[E](infinity)
}

func NewPriorityLimitedPriorityHashQueue[E any](priorityFun func(E) bool) Queue[E] {
	return NewLimitPriorityLimitedPriorityHashQueue(priorityFun, infinity)
}

func NewLimitLimitedPriorityHashQueue[E any](limit int) Queue[E] {
	return NewLimitPriorityLimitedPriorityHashQueue(NoPriority[E], limit)
}

func NewLimitPriorityLimitedPriorityHashQueue[E any](priorityFun func(E) bool, limit int) Queue[E] {
	return newLimitedPriorityHashQueue(priorityFun, limit, false)
}

func NewHashLimitedPriorityHashQueue[E Hashable]() Queue[E] {
	return NewLimitHashLimitedPriorityHashQueue[E](infinity)
}

func NewPriorityHashLimitedPriorityHashQueue[E Hashable](priorityFun func(E) bool) Queue[E] {
	return NewLimitPriorityHashLimitedPriorityHashQueue(priorityFun, infinity)
}

func NewLimitHashLimitedPriorityHashQueue[E Hashable](limit int) Queue[E] {
	return NewLimitPriorityHashLimitedPriorityHashQueue(NoPriority[E], limit)
}

func NewLimitPriorityHashLimitedPriorityHashQueue[E Hashable](priorityFun func(E) bool, limit int) Queue[E] {
	return newLimitedPriorityHashQueue(priorityFun, limit, true)
}

func newLimitedPriorityHashQueue[E any](priorityFun func(E) bool, limit int, hashNeeded bool) Queue[E] {
	var hashMap *shrinkingmap.ShrinkingMap[hashing.HashValue, struct{}]
	if hashNeeded {
		hashMap = shrinkingmap.New[hashing.HashValue, struct{}]()
	}
	return &LimitedPriorityHashQueue[E]{
		deque:       NewLimitedDeque[E](limit),
		pend:        0,
		priorityFun: priorityFun,
		hashMap:     hashMap,
	}
}

// Length returns the number of elements currently stored in the queue.
func (q *LimitedPriorityHashQueue[E]) Length() int {
	return q.deque.Length()
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
func (q *LimitedPriorityHashQueue[E]) Add(elem E) bool {
	var elemHashable Hashable
	var elemHash hashing.HashValue
	var ok bool
	if q.hashMap != nil {
		elemHashable, ok = any(elem).(Hashable)
		if !ok {
			panic("Adding not hashable element")
		}
		elemHash = elemHashable.GetHash()
		if q.hashMap.Has(elemHash) {
			// duplicate element; ignoring
			return false
		}
	}
	priority := q.priorityFun(elem)
	AddFun := func() bool {
		var result bool
		if priority {
			result = q.deque.AddStart(elem)
			if result {
				q.pend++
			}
		} else {
			result = q.deque.AddEnd(elem)
		}
		if result && q.hashMap != nil {
			q.hashMap.Set(elemHash, struct{}{})
		}
		return result
	}
	if AddFun() {
		return true
	}
	if !priority && q.pend == q.Length() {
		// Not possible to add not priority element in queue full of priority elements
		return false
	}
	var deleteElem interface{}
	if q.pend == q.Length() {
		// Queue is full of priority elements and adding priority element: remove last (priority) element
		deleteElem = q.deque.RemoveEnd()
		q.pend--
	} else {
		// There are some non priority elements: delete the oldest non priority element
		deleteElem = q.deque.RemoveAt(q.pend)
	}
	if q.hashMap != nil {
		deleteElemHashable, ok := deleteElem.(Hashable)
		if !ok {
			panic("Deleting not hashable element")
		}
		q.hashMap.Delete(deleteElemHashable.GetHash())
	}
	return AddFun()
}

// Peek returns the element at the head of the queue. This call panics
// if the queue is empty.
func (q *LimitedPriorityHashQueue[E]) Peek() E {
	return q.deque.PeekStart()
}

// Get returns the element at index i in the queue. If the index is
// invalid, the call will panic. This method accepts both positive and
// negative index values. Index 0 refers to the first element, and
// index -1 refers to the last.
func (q *LimitedPriorityHashQueue[E]) Get(i int) E {
	return q.deque.Get(i)
}

// Remove removes and returns the element from the front of the queue. If the
// queue is empty, the call will panic.
func (q *LimitedPriorityHashQueue[E]) Remove() E {
	ret := q.deque.RemoveStart()
	if q.pend > 0 {
		q.pend--
	}
	if q.hashMap != nil {
		retHashable, ok := any(ret).(Hashable)
		if !ok {
			panic("Removing not hashable element")
		}
		q.hashMap.Delete(retHashable.GetHash())
	}
	return ret
}

func NoPriority[E any](E) bool { return false }
