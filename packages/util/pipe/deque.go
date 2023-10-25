package pipe

// dequeImpl is a double sided list, which can be limited in size.
type dequeImpl[E any] struct {
	buf   []E
	head  int // points to the head element
	tail  int // points to the next element after the last one
	count int
	limit int
}

var _ Deque[Hashable] = &dequeImpl[Hashable]{}

const (
	minLen   = 16 // minLen is smallest capacity that deque may have.
	infinity = 0  // used to mark that deque is unlimited
)

// NewDeque creates an unlimited deque
func NewDeque[E any]() Deque[E] {
	return NewLimitedDeque[E](infinity)
}

// NewLimitedDeque creates a deque, which cannot grow above `limit` elements
func NewLimitedDeque[E any](limit int) Deque[E] {
	var initBufSize int
	if (limit != infinity) && (limit < minLen) {
		initBufSize = limit
	} else {
		initBufSize = minLen
	}
	return &dequeImpl[E]{
		head:  0,
		tail:  0,
		count: 0,
		limit: limit,
		buf:   make([]E, initBufSize),
	}
}

// Length returns the number of elements currently stored in the deque.
func (d *dequeImpl[E]) Length() int {
	return d.count
}

func (d *dequeImpl[E]) getIndex(rawIndex int) int {
	index := rawIndex % len(d.buf)
	if index < 0 {
		return index + len(d.buf)
	}
	return index
}

// resize resizes the deque to fit exactly twice its current contents
// this can result in shrinking if the deque is less than half-full
// the size of the resized deque is never smaller than minQueueLen, except
// when the limit is smaller.
func (d *dequeImpl[E]) resize() {
	newSize := d.count << 1
	if newSize < minLen {
		newSize = minLen
	}
	if (d.limit != infinity) && (newSize > d.limit) {
		newSize = d.limit
	}
	newBuf := make([]E, newSize)

	if d.tail > d.head {
		copy(newBuf, d.buf[d.head:d.tail])
	} else {
		n := copy(newBuf, d.buf[d.head:])
		copy(newBuf[n:], d.buf[:d.tail])
	}

	d.head = 0
	d.tail = d.count
	d.buf = newBuf
}

func (d *dequeImpl[E]) prepareAdd() bool {
	if d.count == len(d.buf) {
		if (d.limit != infinity) && (d.count >= d.limit) {
			return false
		}
		d.resize()
	}
	d.count++
	return true
}

// AddStart tries to put an element to the start of the deque. If deque is limited
// and full, returns `false` and doesn't alter the deque. Otherwise successfully
// adds the element and returns `true`.
func (d *dequeImpl[E]) AddStart(elem E) bool {
	if !d.prepareAdd() {
		return false
	}
	d.head = d.getIndex(d.head - 1)
	d.buf[d.head] = elem
	return true
}

// AddEnd tries to put an element to the end of the deque. If deque is limited
// and full, returns `false` and doesn't alter the deque. Otherwise successfully
// adds the element and returns `true`.
func (d *dequeImpl[E]) AddEnd(elem E) bool {
	if !d.prepareAdd() {
		return false
	}
	d.buf[d.tail] = elem
	d.tail = d.getIndex(d.tail + 1)
	return true
}

// PeekStart returns the element at the start of the deque. This call panics
// if the deque is empty.
func (d *dequeImpl[E]) PeekStart() E {
	if d.count <= 0 {
		panic("queue: PeekStart() called on empty queue")
	}
	return d.buf[d.head]
}

// PeekEnd returns the element at the end of the deque. This call panics
// if the deque is empty.
func (d *dequeImpl[E]) PeekEnd() E {
	if d.count <= 0 {
		panic("queue: PeekEnd() called on empty queue")
	}
	return d.buf[d.getIndex(d.tail-1)]
}

// PeekNStart returns a slice of `n` first elements of the deque. The slice is
// independent of the one backing the deque. If there are less than `n` elements
// in the deque, slice of all the elements is returned.
func (d *dequeImpl[E]) PeekNStart(n int) []E {
	if n > d.count {
		n = d.count
	}
	result := make([]E, n)
	leftInTheEnd := len(d.buf) - d.head
	if leftInTheEnd >= n {
		copy(result, d.buf[d.head:d.head+n])
	} else {
		copy(result[:leftInTheEnd], d.buf[d.head:])
		copy(result[leftInTheEnd:], d.buf[:n-leftInTheEnd])
	}
	return result
}

// PeekNEnd returns a slice of `n` last elements of the deque. The slice is
// independent of the one backing the deque. If there are less than `n` elements
// in the deque, slice of all the elements is returned.
func (d *dequeImpl[E]) PeekNEnd(n int) []E {
	if n > d.count {
		n = d.count
	}
	result := make([]E, n)
	if d.tail >= n {
		copy(result, d.buf[d.tail-n:d.tail])
	} else {
		copy(result[:n-d.tail], d.buf[len(d.buf)-n+d.tail:])
		copy(result[n-d.tail:], d.buf[:d.tail])
	}
	return result
}

func (d *dequeImpl[E]) PeekAll() []E {
	return d.PeekNStart(d.Length())
}

func (d *dequeImpl[E]) getAbsoluteIndex(relativeIndex int) int {
	// If indexing backwards, convert to positive index.
	if relativeIndex < 0 {
		relativeIndex += d.count
	}
	if relativeIndex < 0 || relativeIndex >= d.count {
		panic("queue: index out of range")
	}
	return d.getIndex(d.head + relativeIndex)
}

// Get returns the element at index i in the queue. If the index is
// invalid, the call will panic. This method accepts both positive and
// negative index values. Index 0 refers to the first element, and
// index -1 refers to the last.
func (d *dequeImpl[E]) Get(i int) E {
	return d.buf[d.getAbsoluteIndex(i)]
}

func (d *dequeImpl[E]) finaliseRemove() {
	d.count--
	// Resize down if buffer 1/4 full.
	if (len(d.buf) > minLen) && ((d.count << 2) <= len(d.buf)) {
		d.resize()
	}
}

// RemoveStart removes and returns the element from the start of the deque. If the
// deque is empty, the call will panic.
func (d *dequeImpl[E]) RemoveStart() E {
	if d.count <= 0 {
		panic("queue: RemoveStart() called on empty queue")
	}
	ret := d.buf[d.head]
	var nilE E
	d.buf[d.head] = nilE
	d.head = d.getIndex(d.head + 1)
	d.finaliseRemove()
	return ret
}

// RemoveEnd removes and returns the element from the end of the deque. If the
// deque is empty, the call will panic.
func (d *dequeImpl[E]) RemoveEnd() E {
	if d.count <= 0 {
		panic("queue: RemoveEnd() called on empty queue")
	}
	index := d.getIndex(d.tail - 1)
	ret := d.buf[index]
	var nilE E
	d.buf[index] = nilE
	d.tail = index
	d.finaliseRemove()
	return ret
}

// RemoveAt removes and returns `i`th element from the deque. If the index is
// invalid, the call will panic as well as if the deque is empty. This method
// accepts both positive and negative index values. Index 0 refers to the first
// element, and index -1 refers to the last.
func (d *dequeImpl[E]) RemoveAt(i int) E {
	if d.count <= 0 {
		panic("queue: RemoveAt(int) called on empty queue")
	}
	index := d.getAbsoluteIndex(i)
	ret := d.buf[index]
	var nilE E
	if index >= d.head {
		copy(d.buf[d.head+1:index+1], d.buf[d.head:index])
		d.buf[d.head] = nilE
		d.head = d.getIndex(d.head + 1)
	} else {
		copy(d.buf[index:d.tail-1], d.buf[index+1:d.tail])
		d.tail = d.getIndex(d.tail - 1)
		d.buf[d.tail] = nilE
	}
	d.finaliseRemove()
	return ret
}
