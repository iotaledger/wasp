package pipe

// InfinitePipe provides deserialised sender and receiver: it queues messages
// sent by the sender and returns them to the receiver whenever it is ready,
// without blocking the sender process. Depending on the backing queue, the pipe
// might have other characteristics.
type InfinitePipe struct {
	input  chan interface{}
	output chan interface{}
	length chan int
	buffer Queue
}

var _ Pipe = &InfinitePipe{}

func NewDefaultInfinitePipe() Pipe {
	return newInfinitePipe(NewDefaultLimitedPriorityHashQueue())
}

func NewPriorityInfinitePipe(priorityFun func(interface{}) bool) Pipe {
	return newInfinitePipe(NewPriorityLimitedPriorityHashQueue(priorityFun))
}

func NewLimitInfinitePipe(limit int) Pipe {
	return newInfinitePipe(NewLimitLimitedPriorityHashQueue(limit))
}

func NewLimitPriorityInfinitePipe(priorityFun func(interface{}) bool, limit int) Pipe {
	return newInfinitePipe(NewLimitPriorityLimitedPriorityHashQueue(priorityFun, limit))
}

func NewHashInfinitePipe() Pipe {
	return newInfinitePipe(NewHashLimitedPriorityHashQueue(true))
}

func NewPriorityHashInfinitePipe(priorityFun func(interface{}) bool) Pipe {
	return newInfinitePipe(NewPriorityHashLimitedPriorityHashQueue(priorityFun, true))
}

func NewLimitHashInfinitePipe(limit int) Pipe {
	return newInfinitePipe(NewLimitHashLimitedPriorityHashQueue(limit, true))
}

func NewInfinitePipe(priorityFun func(interface{}) bool, limit int) Pipe {
	return newInfinitePipe(NewLimitedPriorityHashQueue(priorityFun, limit, true))
}

func newInfinitePipe(queue Queue) *InfinitePipe {
	ch := &InfinitePipe{
		input:  make(chan interface{}),
		output: make(chan interface{}),
		length: make(chan int),
		buffer: queue,
	}
	go ch.infiniteBuffer()
	return ch
}

func (ch *InfinitePipe) In() chan<- interface{} {
	return ch.input
}

func (ch *InfinitePipe) Out() <-chan interface{} {
	return ch.output
}

func (ch *InfinitePipe) Len() int {
	return <-ch.length
}

func (ch *InfinitePipe) Close() {
	close(ch.input)
}

func (ch *InfinitePipe) infiniteBuffer() {
	var input, output chan interface{}
	var next interface{}
	input = ch.input

	for input != nil || output != nil {
		select {
		case elem, open := <-input:
			if open {
				ch.buffer.Add(elem)
			} else {
				input = nil
			}
		case output <- next:
			ch.buffer.Remove()
		case ch.length <- ch.buffer.Length():
		}

		if ch.buffer.Length() > 0 {
			output = ch.output
			next = ch.buffer.Peek()
		} else {
			output = nil
			next = nil
		}
	}

	close(ch.output)
	close(ch.length)
}
