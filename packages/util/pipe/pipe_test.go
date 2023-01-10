package pipe

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInfinitePipeWriteReadLen(t *testing.T) {
	testDefaultPipeWriteReadLen(NewSimpleNothashableFactory(), NewInfinitePipe[SimpleNothashable](), 1000, identityFunInt, t)
}

func TestPriorityInfinitePipeWriteReadLen(t *testing.T) {
	testPriorityPipeWriteReadLen(NewSimpleNothashableFactory(), NewPriorityInfinitePipe[SimpleNothashable], t)
}

func TestLimitInfinitePipeNoLimitWriteReadLen(t *testing.T) {
	testLimitedPipeNoLimitWriteReadLen(NewSimpleNothashableFactory(), NewLimitInfinitePipe[SimpleNothashable], t)
}

func TestLimitInfinitePipeWriteReadLen(t *testing.T) {
	testLimitedPipeWriteReadLen(NewSimpleNothashableFactory(), NewLimitInfinitePipe[SimpleNothashable], t)
}

func TestLimitPriorityInfinitePipeNoLimitWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeNoLimitWriteReadLen(NewSimpleNothashableFactory(), NewLimitPriorityInfinitePipe[SimpleNothashable], t)
}

func TestLimitPriorityInfinitePipeWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeWriteReadLen(NewSimpleNothashableFactory(), NewLimitPriorityInfinitePipe[SimpleNothashable], t)
}

func TestHashInfinitePipeWriteReadLen(t *testing.T) {
	testDefaultPipeWriteReadLen(NewSimpleHashableFactory(), NewHashInfinitePipe[SimpleHashable](), 1000, identityFunInt, t)
}

func TestPriorityHashInfinitePipeWriteReadLen(t *testing.T) {
	testPriorityPipeWriteReadLen(NewSimpleHashableFactory(), NewPriorityHashInfinitePipe[SimpleHashable], t)
}

func TestLimitHashInfinitePipeNoLimitWriteReadLen(t *testing.T) {
	testLimitedPipeNoLimitWriteReadLen(NewSimpleHashableFactory(), NewLimitHashInfinitePipe[SimpleHashable], t)
}

func TestLimitHashInfinitePipeWriteReadLen(t *testing.T) {
	testLimitedPipeWriteReadLen(NewSimpleHashableFactory(), NewLimitHashInfinitePipe[SimpleHashable], t)
}

func TestLimitPriorityHashInfinitePipeNoLimitWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeNoLimitWriteReadLen(NewSimpleHashableFactory(), NewLimitPriorityHashInfinitePipe[SimpleHashable], t)
}

func TestLimitPriorityHashInfinitePipeWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeWriteReadLen(NewSimpleHashableFactory(), NewLimitPriorityHashInfinitePipe[SimpleHashable], t)
}

func testLimitedPriorityPipeNoLimitWriteReadLen[E IntConvertable](factory Factory[E], makeLimitedPriorityPipeFun func(priorityFun func(E) bool, limit int) Pipe[E], t *testing.T) {
	testPriorityPipeWriteReadLen(factory, func(priorityFun func(E) bool) Pipe[E] {
		return makeLimitedPriorityPipeFun(priorityFun, 1200)
	}, t)
}

func testLimitedPriorityPipeWriteReadLen[E IntConvertable](factory Factory[E], makeLimitedPriorityPipeFun func(priorityFun func(E) bool, limit int) Pipe[E], t *testing.T) {
	limit := 800
	p := makeLimitedPriorityPipeFun(priorityFunMod3[E], limit)
	result := func(index int) int {
		if index <= 333 {
			return -3*index + 999
		}
		if index%2 == 0 {
			return 3*index/2 - 200
		}
		return (3*index - 401) / 2
	}
	testPipeWriteReadLen(factory, p, 1000, limit, result, t)
}

func testLimitedPipeNoLimitWriteReadLen[E IntConvertable](factory Factory[E], makeLimitedPipeFun func(limit int) Pipe[E], t *testing.T) {
	testDefaultPipeWriteReadLen(factory, makeLimitedPipeFun(1200), 1000, identityFunInt, t)
}

func testLimitedPipeWriteReadLen[E IntConvertable](factory Factory[E], makeLimitedPipeFun func(limit int) Pipe[E], t *testing.T) {
	limit := 800
	elementsToAdd := 1000
	indexDiff := elementsToAdd - limit
	result := func(index int) int {
		return index + indexDiff
	}
	testPipeWriteReadLen(factory, makeLimitedPipeFun(limit), elementsToAdd, limit, result, t)
}

func testPriorityPipeWriteReadLen[E IntConvertable](factory Factory[E], makePriorityPipeFun func(func(E) bool) Pipe[E], t *testing.T) {
	p := makePriorityPipeFun(priorityFunMod3[E])
	result := func(index int) int {
		if index <= 333 {
			return -3*index + 999
		}
		if index%2 == 0 {
			return 3*index/2 - 500
		}
		return (3*index - 1001) / 2
	}
	testDefaultPipeWriteReadLen(factory, p, 1000, result, t)
}

func testDefaultPipeWriteReadLen[E IntConvertable](factory Factory[E], p Pipe[E], elementsToWrite int, result func(index int) int, t *testing.T) {
	testPipeWriteReadLen(factory, p, elementsToWrite, elementsToWrite, result, t)
}

func testPipeWriteReadLen[E IntConvertable](factory Factory[E], p Pipe[E], elementsToWrite, elementsToRead int, result func(index int) int, t *testing.T) {
	for i := 0; i < elementsToWrite; i++ {
		p.In() <- factory.Create(i)
	}
	fullLength := p.Len()
	require.Equalf(t, elementsToRead, fullLength, "full channel length mismatch")
	p.Close()
	closedLength := p.Len()
	require.Equalf(t, elementsToRead, closedLength, "closed channel length mismatch")
	for i := 0; i < elementsToRead; i++ {
		val := <-p.Out()
		require.Equalf(t, factory.Create(result(i)), val, "read %d mismatch", i)
	}
}

//--

func TestInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	result := identityFunInt
	testDefaultPipeConcurrentWriteReadLen(NewSimpleNothashableFactory(), NewInfinitePipe[SimpleNothashable](), 1000, &result, t)
}

func TestPriorityInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testPriorityPipeConcurrentWriteReadLen(NewSimpleNothashableFactory(), NewPriorityInfinitePipe[SimpleNothashable], t)
}

func TestLimitInfinitePipeNoLimitConcurrentWriteReadLen(t *testing.T) {
	testLimitedPipeNoLimitConcurrentWriteReadLen(NewSimpleNothashableFactory(), NewLimitInfinitePipe[SimpleNothashable], t)
}

func TestLimitInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testLimitedPipeConcurrentWriteReadLen(NewSimpleNothashableFactory(), NewLimitInfinitePipe[SimpleNothashable], t)
}

func TestLimitPriorityInfinitePipeNoLimitConcurrentWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeNoLimitConcurrentWriteReadLen(NewSimpleNothashableFactory(), NewLimitPriorityInfinitePipe[SimpleNothashable], t)
}

func TestLimitPriorityInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeConcurrentWriteReadLen(NewSimpleNothashableFactory(), NewLimitPriorityInfinitePipe[SimpleNothashable], t)
}

func TestHashInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	result := identityFunInt
	testDefaultPipeConcurrentWriteReadLen(NewSimpleHashableFactory(), NewHashInfinitePipe[SimpleHashable](), 1000, &result, t)
}

func TestPriorityHashInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testPriorityPipeConcurrentWriteReadLen(NewSimpleHashableFactory(), NewPriorityHashInfinitePipe[SimpleHashable], t)
}

func TestLimitHashInfinitePipeNoLimitConcurrentWriteReadLen(t *testing.T) {
	testLimitedPipeNoLimitConcurrentWriteReadLen(NewSimpleHashableFactory(), NewLimitHashInfinitePipe[SimpleHashable], t)
}

func TestLimitHashInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testLimitedPipeConcurrentWriteReadLen(NewSimpleHashableFactory(), NewLimitHashInfinitePipe[SimpleHashable], t)
}

func TestLimitPriorityHashInfinitePipeNoLimitConcurrentWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeNoLimitConcurrentWriteReadLen(NewSimpleHashableFactory(), NewLimitPriorityHashInfinitePipe[SimpleHashable], t)
}

func TestLimitPriorityHashInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeConcurrentWriteReadLen(NewSimpleHashableFactory(), NewLimitPriorityHashInfinitePipe[SimpleHashable], t)
}

func testLimitedPriorityPipeNoLimitConcurrentWriteReadLen[E IntConvertable](factory Factory[E], makeLimitedPriorityPipeFun func(priorityFun func(E) bool, limit int) Pipe[E], t *testing.T) {
	testPriorityPipeConcurrentWriteReadLen(factory, func(priorityFun func(E) bool) Pipe[E] {
		return makeLimitedPriorityPipeFun(priorityFun, 1200)
	}, t)
}

func testLimitedPriorityPipeConcurrentWriteReadLen[E IntConvertable](factory Factory[E], makeLimitedPriorityPipeFun func(priorityFun func(E) bool, limit int) Pipe[E], t *testing.T) {
	limit := 800
	ch := makeLimitedPriorityPipeFun(priorityFunMod3[E], limit)
	testPipeConcurrentWriteReadLen(factory, ch, 1000, limit, nil, t)
}

func testLimitedPipeNoLimitConcurrentWriteReadLen[E IntConvertable](factory Factory[E], makeLimitedPipeFun func(limit int) Pipe[E], t *testing.T) {
	result := identityFunInt
	testDefaultPipeConcurrentWriteReadLen(factory, makeLimitedPipeFun(1200), 1000, &result, t)
}

func testLimitedPipeConcurrentWriteReadLen[E IntConvertable](factory Factory[E], makeLimitedPipeFun func(limit int) Pipe[E], t *testing.T) {
	testPipeConcurrentWriteReadLen(factory, makeLimitedPipeFun(800), 1000, 800, nil, t)
}

func testPriorityPipeConcurrentWriteReadLen[E IntConvertable](factory Factory[E], makePriorityPipeFun func(func(E) bool) Pipe[E], t *testing.T) {
	ch := makePriorityPipeFun(priorityFunMod3[E])
	testDefaultPipeConcurrentWriteReadLen(factory, ch, 1000, nil, t)
}

func testDefaultPipeConcurrentWriteReadLen[E IntConvertable](factory Factory[E], p Pipe[E], elementsToWrite int, result *func(index int) int, t *testing.T) {
	testPipeConcurrentWriteReadLen(factory, p, elementsToWrite, elementsToWrite, result, t)
}

func testPipeConcurrentWriteReadLen[E IntConvertable](factory Factory[E], p Pipe[E], elementsToWrite, elementsToRead int, result *func(index int) int, t *testing.T) {
	var wg sync.WaitGroup
	written := 0
	read := 0
	stop := make(chan bool)
	wg.Add(2)

	go func() {
		for i := 0; i < elementsToWrite; i++ {
			p.In() <- factory.Create(i)
			written++
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < elementsToRead; i++ {
			val := <-p.Out()
			if result != nil {
				require.Equalf(t, factory.Create((*result)(i)), val, "concurrent read %d mismatch", i)
			}
			read++
		}
		wg.Done()
	}()

	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				length := p.Len()
				t.Logf("current channel length is %d", length)
				// no asserts here - the read/write process is asynchronous
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	wg.Wait()
	stop <- true
	require.Equalf(t, elementsToWrite, written, "concurrent write elements written mismatch")
	require.Equalf(t, elementsToRead, read, "concurrent read elements read mismatch")
}
