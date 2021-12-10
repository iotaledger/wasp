package pipe

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultInfinitePipeWriteReadLen(t *testing.T) {
	testDefaultPipeWriteReadLen(NewDefaultInfinitePipe(), 1000, identityFunInt, t)
}

func TestPriorityInfinitePipeWriteReadLen(t *testing.T) {
	testPriorityPipeWriteReadLen(NewPriorityInfinitePipe, t)
}

func TestLimitInfinitePipeNoLimitWriteReadLen(t *testing.T) {
	testLimitedPipeNoLimitWriteReadLen(NewLimitInfinitePipe, t)
}

func TestLimitInfinitePipeWriteReadLen(t *testing.T) {
	testLimitedPipeWriteReadLen(NewLimitInfinitePipe, t)
}

func TestLimitPriorityInfinitePipeNoLimitWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeNoLimitWriteReadLen(NewLimitPriorityInfinitePipe, t)
}

func TestLimitPriorityInfinitePipeWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeWriteReadLen(NewLimitPriorityInfinitePipe, t)
}

func TestHashInfinitePipeWriteReadLen(t *testing.T) {
	testDefaultPipeWriteReadLen(NewHashInfinitePipe(), 1000, identityFunInt, t)
}

func TestPriorityHashInfinitePipeWriteReadLen(t *testing.T) {
	testPriorityPipeWriteReadLen(NewPriorityHashInfinitePipe, t)
}

func TestLimitHashInfinitePipeNoLimitWriteReadLen(t *testing.T) {
	testLimitedPipeNoLimitWriteReadLen(NewLimitHashInfinitePipe, t)
}

func TestLimitHashInfinitePipeWriteReadLen(t *testing.T) {
	testLimitedPipeWriteReadLen(NewLimitHashInfinitePipe, t)
}

func TestInfinitePipeNoLimitWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeNoLimitWriteReadLen(NewInfinitePipe, t)
}

func TestInfinitePipeWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeWriteReadLen(NewInfinitePipe, t)
}

func testLimitedPriorityPipeNoLimitWriteReadLen(makeLimitedPriorityPipeFun func(priorityFun func(i interface{}) bool, limit int) Pipe, t *testing.T) {
	testPriorityPipeWriteReadLen(func(priorityFun func(i interface{}) bool) Pipe { return makeLimitedPriorityPipeFun(priorityFun, 1200) }, t)
}

func testLimitedPriorityPipeWriteReadLen(makeLimitedPriorityPipeFun func(priorityFun func(i interface{}) bool, limit int) Pipe, t *testing.T) {
	limit := 800
	p := makeLimitedPriorityPipeFun(priorityFunMod3, limit)
	result := func(index int) int {
		if index <= 333 {
			return -3*index + 999
		}
		if index%2 == 0 {
			return 3*index/2 - 200
		}
		return (3*index - 401) / 2
	}
	testPipeWriteReadLen(p, 1000, limit, result, t)
}

func testLimitedPipeNoLimitWriteReadLen(makeLimitedPipeFun func(limit int) Pipe, t *testing.T) {
	testDefaultPipeWriteReadLen(makeLimitedPipeFun(1200), 1000, identityFunInt, t)
}

func testLimitedPipeWriteReadLen(makeLimitedPipeFun func(limit int) Pipe, t *testing.T) {
	limit := 800
	elementsToAdd := 1000
	indexDiff := elementsToAdd - limit
	result := func(index int) int {
		return index + indexDiff
	}
	testPipeWriteReadLen(makeLimitedPipeFun(limit), elementsToAdd, limit, result, t)
}

func testPriorityPipeWriteReadLen(makePriorityPipeFun func(func(i interface{}) bool) Pipe, t *testing.T) {
	p := makePriorityPipeFun(priorityFunMod3)
	result := func(index int) int {
		if index <= 333 {
			return -3*index + 999
		}
		if index%2 == 0 {
			return 3*index/2 - 500
		}
		return (3*index - 1001) / 2
	}
	testDefaultPipeWriteReadLen(p, 1000, result, t)
}

func testDefaultPipeWriteReadLen(p Pipe, elementsToWrite int, result func(index int) int, t *testing.T) {
	testPipeWriteReadLen(p, elementsToWrite, elementsToWrite, result, t)
}

func testPipeWriteReadLen(p Pipe, elementsToWrite, elementsToRead int, result func(index int) int, t *testing.T) {
	for i := 0; i < elementsToWrite; i++ {
		p.In() <- SimpleHashable(i)
	}
	fullLength := p.Len()
	require.Equalf(t, elementsToRead, fullLength, "full channel length missmatch")
	p.Close()
	closedLength := p.Len()
	require.Equalf(t, elementsToRead, closedLength, "closed channel length missmatch")
	for i := 0; i < elementsToRead; i++ {
		val := <-p.Out()
		require.Equalf(t, SimpleHashable(result(i)), val.(SimpleHashable), "read %d missmatch", i)
	}
}

//--

func TestDefaultInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	result := identityFunInt
	testDefaultPipeConcurrentWriteReadLen(NewDefaultInfinitePipe(), 1000, &result, t)
}

func TestPriorityInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testPriorityPipeConcurrentWriteReadLen(NewPriorityInfinitePipe, t)
}

func TestLimitInfinitePipeNoLimitConcurrentWriteReadLen(t *testing.T) {
	testLimitedPipeNoLimitConcurrentWriteReadLen(NewLimitInfinitePipe, t)
}

func TestLimitInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testLimitedPipeConcurrentWriteReadLen(NewLimitInfinitePipe, t)
}

func TestLimitPriorityInfinitePipeNoLimitConcurrentWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeNoLimitConcurrentWriteReadLen(NewLimitPriorityInfinitePipe, t)
}

func TestLimitPriorityInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeConcurrentWriteReadLen(NewLimitPriorityInfinitePipe, t)
}

func TestHashInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	result := identityFunInt
	testDefaultPipeConcurrentWriteReadLen(NewHashInfinitePipe(), 1000, &result, t)
}

func TestPriorityHashInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testPriorityPipeConcurrentWriteReadLen(NewPriorityHashInfinitePipe, t)
}

func TestLimitHashInfinitePipeNoLimitConcurrentWriteReadLen(t *testing.T) {
	testLimitedPipeNoLimitConcurrentWriteReadLen(NewLimitHashInfinitePipe, t)
}

func TestLimitHashInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testLimitedPipeConcurrentWriteReadLen(NewLimitHashInfinitePipe, t)
}

func TestInfinitePipeNoLimitConcurrentWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeNoLimitConcurrentWriteReadLen(NewInfinitePipe, t)
}

func TestInfinitePipeConcurrentWriteReadLen(t *testing.T) {
	testLimitedPriorityPipeConcurrentWriteReadLen(NewInfinitePipe, t)
}

func testLimitedPriorityPipeNoLimitConcurrentWriteReadLen(makeLimitedPriorityPipeFun func(priorityFun func(i interface{}) bool, limit int) Pipe, t *testing.T) {
	testPriorityPipeConcurrentWriteReadLen(func(priorityFun func(i interface{}) bool) Pipe { return makeLimitedPriorityPipeFun(priorityFun, 1200) }, t)
}

func testLimitedPriorityPipeConcurrentWriteReadLen(makeLimitedPriorityPipeFun func(priorityFun func(i interface{}) bool, limit int) Pipe, t *testing.T) {
	limit := 800
	ch := makeLimitedPriorityPipeFun(priorityFunMod3, limit)
	testPipeConcurrentWriteReadLen(ch, 1000, limit, nil, t)
}

func testLimitedPipeNoLimitConcurrentWriteReadLen(makeLimitedPipeFun func(limit int) Pipe, t *testing.T) {
	result := identityFunInt
	testDefaultPipeConcurrentWriteReadLen(makeLimitedPipeFun(1200), 1000, &result, t)
}

func testLimitedPipeConcurrentWriteReadLen(makeLimitedPipeFun func(limit int) Pipe, t *testing.T) {
	testPipeConcurrentWriteReadLen(makeLimitedPipeFun(800), 1000, 800, nil, t)
}

func testPriorityPipeConcurrentWriteReadLen(makePriorityPipeFun func(func(i interface{}) bool) Pipe, t *testing.T) {
	ch := makePriorityPipeFun(priorityFunMod3)
	testDefaultPipeConcurrentWriteReadLen(ch, 1000, nil, t)
}

func testDefaultPipeConcurrentWriteReadLen(p Pipe, elementsToWrite int, result *func(index int) int, t *testing.T) {
	testPipeConcurrentWriteReadLen(p, elementsToWrite, elementsToWrite, result, t)
}

func testPipeConcurrentWriteReadLen(p Pipe, elementsToWrite, elementsToRead int, result *func(index int) int, t *testing.T) {
	var wg sync.WaitGroup
	written := 0
	read := 0
	stop := make(chan bool)
	wg.Add(2)

	go func() {
		for i := 0; i < elementsToWrite; i++ {
			p.In() <- SimpleHashable(i)
			written++
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < elementsToRead; i++ {
			val := <-p.Out()
			if result != nil {
				require.Equalf(t, SimpleHashable((*result)(i)), val.(SimpleHashable), "concurent read %d missmatch", i)
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
				// no asserts here - the read/write process is asynchronious
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	wg.Wait()
	stop <- true
	require.Equalf(t, elementsToWrite, written, "concurent write elements written missmatch")
	require.Equalf(t, elementsToRead, read, "concurent read elements read missmatch")
}
