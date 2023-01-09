package pipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testQueueBasicAddLengthPeekRemove[E IntBased](factory Factory[E], q Queue[E], elementsToAdd int, add func(index int) int, addResult func(index int) bool, elementsToRemove int, result func(index int) int, t *testing.T) {
	for i := 0; i < elementsToAdd; i++ {
		value := factory.Create(add(i))
		actualAddResult := q.Add(value)
		require.Equalf(t, addResult(i), actualAddResult, "add result of element %d value %d mismatch", i, value)
	}
	fullLength := q.Length()
	require.Equalf(t, elementsToRemove, fullLength, "full queue length mismatch")
	for i := 0; i < elementsToRemove; i++ {
		expected := factory.Create(result(i))
		peekResult := q.Peek()
		require.Equalf(t, expected, peekResult, "peek %d mismatch", i)
		removeResult := q.Remove()
		require.Equalf(t, expected, removeResult, "remove %d mismatch", i)
	}
	emptyLength := q.Length()
	require.Equalf(t, 0, emptyLength, "empty queue length mismatch")
}

//--

func TestLimitedPriorityHashQueueSimple(t *testing.T) {
	testDefaultQueueSimple(NewSimpleNothashableFactory(), NewLimitedPriorityHashQueue[SimpleNothashable](), t)
}

func TestPriorityLimitedPriorityHashQueueSimple(t *testing.T) {
	testPriorityQueueSimple(NewSimpleNothashableFactory(), NewPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedQueueNoLimitSimple(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedQueueSimple(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedPriorityQueueNoLimitSimple(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedPriorityQueueSimple(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestHashLimitedPriorityHashQueueSimple(t *testing.T) {
	testDefaultQueueSimple(NewSimpleHashableFactory(), NewHashLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityHashLimitedPriorityHashQueueSimple(t *testing.T) {
	testPriorityQueueSimple(NewSimpleHashableFactory(), NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedQueueNoLimitSimple(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedQueueSimple(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedPriorityQueueNoLimitSimple(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedPriorityQueueSimple(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func testLimitedPriorityQueueNoLimitSimple[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(e E) bool, limit int) Queue[E], t *testing.T) {
	testPriorityQueueSimple(factory, func(priorityFun func(e E) bool) Queue[E] {
		return makeLimitedPriorityQueueFun(priorityFun, 15)
	}, t)
}

func testLimitedPriorityQueueSimple[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(e E) bool, limit int) Queue[E], t *testing.T) {
	resultArray := []int{9, 6, 3, 0, 4, 5, 7, 8}
	limit := len(resultArray)
	q := makeLimitedPriorityQueueFun(priorityFunMod3[E], limit)
	result := func(index int) int {
		return resultArray[index]
	}
	testQueueSimple(factory, q, 10, limit, result, t)
}

func testLimitedQueueNoLimitSimple[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	testDefaultQueueSimple(factory, makeLimitedQueueFun(15), t)
}

func testLimitedQueueSimple[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	limit := 8
	elementsToAdd := 10
	indexDiff := elementsToAdd - limit
	q := makeLimitedQueueFun(limit)
	result := func(index int) int {
		return index + indexDiff
	}
	testQueueSimple(factory, q, elementsToAdd, limit, result, t)
}

func testPriorityQueueSimple[E IntBased](factory Factory[E], makePriorityQueueFun func(func(e E) bool) Queue[E], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3[E])
	resultArray := []int{9, 6, 3, 0, 1, 2, 4, 5, 7, 8}
	result := func(index int) int {
		return resultArray[index]
	}
	elementsToAdd := len(resultArray)
	testQueueSimple(factory, q, elementsToAdd, elementsToAdd, result, t)
}

func testDefaultQueueSimple[E IntBased](factory Factory[E], q Queue[E], t *testing.T) {
	elementsToAdd := 10
	testQueueSimple(factory, q, elementsToAdd, elementsToAdd, identityFunInt, t)
}

func testQueueSimple[E IntBased](factory Factory[E], q Queue[E], elementsToAdd, elementsToRemove int, result func(index int) int, t *testing.T) {
	testQueueBasicAddLengthPeekRemove(factory, q, elementsToAdd, identityFunInt, alwaysTrueFun, elementsToRemove, result, t)
}

//--

func TestLimitedPriorityHashQueueTwice(t *testing.T) {
	testDefaultQueueTwice(NewSimpleNothashableFactory(), NewLimitedPriorityHashQueue[SimpleNothashable](), t)
}

func TestPriorityLimitedPriorityHashQueueTwice(t *testing.T) {
	testPriorityQueueTwice(NewSimpleNothashableFactory(), NewPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testDefaultQueueTwice(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable](150), t)
}

func TestLimitLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 80
	elementsToAddSingle := 50
	indexDiff := 2*elementsToAddSingle - limit
	q := NewLimitLimitedPriorityHashQueue[SimpleNothashable](limit)
	resultFun := func(index int) int {
		return (index + indexDiff) % elementsToAddSingle
	}
	testQueueTwice(NewSimpleNothashableFactory(), q, elementsToAddSingle, alwaysTrueFun, limit, resultFun, t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testPriorityQueueTwice(NewSimpleNothashableFactory(), func(priorityFun func(i SimpleNothashable) bool) Queue[SimpleNothashable] {
		return NewLimitPriorityLimitedPriorityHashQueue(priorityFun, 150)
	}, t)
}

func TestLimitPriorityLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 80
	elementsToAddSingle := 50
	q := NewLimitPriorityLimitedPriorityHashQueue(priorityFunMod3[SimpleNothashable], limit)
	resultFun := func(index int) int {
		if index <= 16 {
			return 48 - 3*index
		} else if index <= 33 {
			return 99 - 3*index
		} else if index <= 46 {
			if index%2 == 0 {
				return 3*index/2 - 20
			}
			return (3*index - 41) / 2
		} else {
			if index%2 == 1 {
				return (3*index - 139) / 2
			}
			return 3*index/2 - 70
		}
	}
	testQueueTwice(NewSimpleNothashableFactory(), q, elementsToAddSingle, alwaysTrueFun, limit, resultFun, t)
}

func TestHashLimitedPriorityHashQueueTwice(t *testing.T) {
	testHashQueueTwice(NewHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestPriorityHashLimitedPriorityHashQueueTwice(t *testing.T) {
	testPriorityHashQueueTwice(NewSimpleHashableFactory(), NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testHashQueueTwice(func() Queue[SimpleHashable] {
		return NewLimitHashLimitedPriorityHashQueue[SimpleHashable](80)
	}, t)
}

func TestLimitHashLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 30
	elementsToAddSingle := 50
	indexDiff := elementsToAddSingle - limit
	resultFun := func(index int) int { return index + indexDiff }
	q := NewLimitHashLimitedPriorityHashQueue[SimpleHashable](limit)
	testQueueTwice(NewSimpleHashableFactory(), q, elementsToAddSingle, alwaysTrueFun, limit, resultFun, t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testPriorityHashQueueTwice(NewSimpleHashableFactory(), func(priorityFun func(i SimpleHashable) bool) Queue[SimpleHashable] {
		return NewLimitPriorityHashLimitedPriorityHashQueue(priorityFun, 80)
	}, t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 30
	elementsToAddSingle := 50
	q := NewLimitPriorityHashLimitedPriorityHashQueue(priorityFunMod3[SimpleHashable], limit)
	addResultFun := func(index int) bool { return (index < elementsToAddSingle) || ((index-elementsToAddSingle)%3 != 0) }
	resultFun := func(index int) int {
		if index <= 16 {
			return 48 - 3*index
		}
		if index%2 == 1 {
			return (3*index + 11) / 2
		}
		return 3*index/2 + 5
	}
	testQueueTwice(NewSimpleHashableFactory(), q, elementsToAddSingle, addResultFun, limit, resultFun, t)
}

func testHashQueueTwice(makeHashQueueFun func() Queue[SimpleHashable], t *testing.T) {
	q := makeHashQueueFun()
	elementsToAddSingle := 50
	addResultFun := func(index int) bool { return index < elementsToAddSingle }
	testQueueTwice(NewSimpleHashableFactory(), q, elementsToAddSingle, addResultFun, elementsToAddSingle, identityFunInt, t)
}

func testPriorityHashQueueTwice[E IntBased](factory Factory[E], makePriorityHashQueueFun func(priorityFun func(E) bool) Queue[E], t *testing.T) {
	q := makePriorityHashQueueFun(priorityFunMod3[E])
	elementsToAddSingle := 50
	addResultFun := func(index int) bool { return index < elementsToAddSingle }
	resultFun := func(index int) int {
		if index <= 16 {
			return 48 - 3*index
		}
		if index%2 == 1 {
			return (3*index - 49) / 2
		}
		return 3*index/2 - 25
	}
	testQueueTwice(factory, q, elementsToAddSingle, addResultFun, elementsToAddSingle, resultFun, t)
}

func testPriorityQueueTwice[E IntBased](factory Factory[E], makePriorityQueueFun func(func(E) bool) Queue[E], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3[E])
	elementsToAddSingle := 50
	resultFun := func(index int) int {
		if index <= 16 {
			return 48 - 3*index
		} else if index <= 33 {
			return 99 - 3*index
		} else if index <= 66 {
			if index%2 == 0 {
				return 3*index/2 - 50
			}
			return (3*index - 101) / 2
		} else {
			if index%2 == 1 {
				return (3*index - 199) / 2
			}
			return 3*index/2 - 100
		}
	}
	testQueueTwice(factory, q, elementsToAddSingle, alwaysTrueFun, 2*elementsToAddSingle, resultFun, t)
}

func testDefaultQueueTwice[E IntBased](factory Factory[E], q Queue[E], t *testing.T) {
	elementsToAddSingle := 50
	resultFun := func(index int) int { return index % elementsToAddSingle }
	testQueueTwice(factory, q, elementsToAddSingle, alwaysTrueFun, 2*elementsToAddSingle, resultFun, t)
}

func testQueueTwice[E IntBased](factory Factory[E], q Queue[E], elementsToAddSingle int, addResult func(index int) bool, elementsToRemove int, result func(index int) int, t *testing.T) {
	addFun := func(index int) int {
		return index % elementsToAddSingle
	}
	testQueueBasicAddLengthPeekRemove(factory, q, 2*elementsToAddSingle, addFun, addResult, elementsToRemove, result, t)
}

//--

func TestLimitPriorityLimitedPriorityHashQueueOverflow(t *testing.T) {
	factory := NewSimpleNothashableFactory()
	limit := 30
	elementsToAddSingle := 50
	cutOff := elementsToAddSingle / 2
	cutOffSh := factory.Create(cutOff)
	q := NewLimitPriorityLimitedPriorityHashQueue(func(e SimpleNothashable) bool {
		return e < cutOffSh
	}, limit)
	addResultFun := func(index int) bool {
		return index < elementsToAddSingle+cutOff
	}
	resultFun := func(index int) int {
		if index < 25 {
			return 24 - index
		}
		return 49 - index
	}
	testQueueTwice(factory, q, elementsToAddSingle, addResultFun, limit, resultFun, t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueOverflow(t *testing.T) {
	factory := NewSimpleHashableFactory()
	limit := 30
	elementsToAddSingle := 50
	cutOffLow := factory.Create(20)
	cutOffHigh := factory.Create(40)
	q := NewLimitPriorityHashLimitedPriorityHashQueue(func(e SimpleHashable) bool {
		return e < cutOffLow || cutOffHigh <= e
	}, limit)
	addResultFun := func(index int) bool {
		return index < elementsToAddSingle
	}
	resultFun := func(index int) int {
		if index < 10 {
			return 49 - index
		}
		return 29 - index
	}
	testQueueTwice(factory, q, elementsToAddSingle, addResultFun, limit, resultFun, t)
}

//--

func TestLimitPriorityHashLimitedPriorityHashQueueDuplicates(t *testing.T) {
	limit := 80
	elementsToAddFirstIteration := 50
	q := NewLimitPriorityHashLimitedPriorityHashQueue(priorityFunMod3[SimpleHashable], limit)
	addFun := func(index int) int {
		if index < elementsToAddFirstIteration {
			return 2 * index
		}
		return index - elementsToAddFirstIteration
	}
	addResultFun := func(index int) bool {
		return (index < elementsToAddFirstIteration) || ((index-elementsToAddFirstIteration)%2 == 1)
	}
	resultFun := func(index int) int {
		if index <= 16 {
			return 99 - 6*index
		} else if index <= 33 {
			return 198 - 6*index
		} else if index <= 46 {
			if index%2 == 0 {
				return 3*index - 40
			}
			return 3*index - 41
		} else {
			if index%2 == 0 {
				return 3*index - 139
			}
			return 3*index - 140
		}
	}
	testQueueBasicAddLengthPeekRemove(NewSimpleHashableFactory(), q, 3*elementsToAddFirstIteration, addFun, addResultFun, limit, resultFun, t)
}

//--

func TestLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testDefaultQueueAddRemove(NewSimpleNothashableFactory(), NewLimitedPriorityHashQueue[SimpleNothashable](), t)
}

func TestPriorityLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testPriorityQueueAddRemove(NewSimpleNothashableFactory(), NewPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedQueueNoLimitAddRemove(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedQueueAddRemove(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedPriorityQueueNoLimitAddRemove(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedPriorityQueueAddRemove(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestHashLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testDefaultQueueAddRemove(NewSimpleHashableFactory(), NewHashLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityHashLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testPriorityQueueAddRemove(NewSimpleHashableFactory(), NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedQueueNoLimitAddRemove(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedQueueAddRemove(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedPriorityQueueNoLimitAddRemove(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedPriorityQueueAddRemove(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func testLimitedPriorityQueueNoLimitAddRemove[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	testPriorityQueueAddRemove(factory, func(priorityFun func(E) bool) Queue[E] {
		return makeLimitedPriorityQueueFun(priorityFun, 150)
	}, t)
}

func testLimitedPriorityQueueAddRemove[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	limit := 80
	q := makeLimitedPriorityQueueFun(priorityFunMod3[E], limit)
	result := func(index int) int {
		if index%2 == 0 {
			return 3*index/2 + 31
		}
		return (3*index + 61) / 2
	}
	testQueueAddRemove(factory, q, 100, 50, limit, result, t)
}

func testLimitedQueueNoLimitAddRemove[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	testDefaultQueueAddRemove(factory, makeLimitedQueueFun(150), t)
}

func testLimitedQueueAddRemove[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	limit := 80
	elementsToAdd := 100
	elementsToRemoveAdd := 50
	indexDiff := elementsToAdd - limit + elementsToRemoveAdd
	q := makeLimitedQueueFun(limit)
	result := func(index int) int {
		return index + indexDiff
	}
	testQueueAddRemove(factory, q, elementsToAdd, elementsToRemoveAdd, limit, result, t)
}

func testPriorityQueueAddRemove[E IntBased](factory Factory[E], makePriorityQueueFun func(func(E) bool) Queue[E], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3[E])
	result := func(index int) int {
		if index%2 == 0 {
			return 3*index/2 + 1
		}
		return (3*index + 1) / 2
	}
	elementsToAdd := 100
	testQueueAddRemove(factory, q, elementsToAdd, 50, elementsToAdd, result, t)
}

func testDefaultQueueAddRemove[E IntBased](factory Factory[E], q Queue[E], t *testing.T) {
	elementsToAdd := 100
	elementsToRemoveAdd := 50
	testQueueAddRemove(factory, q, elementsToAdd, elementsToRemoveAdd, elementsToAdd, func(index int) int { return index + elementsToRemoveAdd }, t)
}

func testQueueAddRemove[E IntBased](factory Factory[E], q Queue[E], elementsToAdd, elementsToRemoveAdd, elementsToRemove int, result func(index int) int, t *testing.T) {
	for i := 0; i < elementsToAdd; i++ {
		require.Truef(t, q.Add(factory.Create(i)), "failed to add element %d", i)
	}
	for i := 0; i < elementsToRemoveAdd; i++ {
		q.Remove()
		add := elementsToAdd + i
		require.Truef(t, q.Add(factory.Create(add)), "failed to add element %d", add)
	}
	fullLength := q.Length()
	require.Equalf(t, elementsToRemove, fullLength, "full queue length mismatch")

	for i := 0; i < elementsToRemove; i++ {
		expected := factory.Create(result(i))
		peekResult := q.Peek()
		require.Equalf(t, expected, peekResult, "peek %d mismatch", i)
		removeResult := q.Remove()
		require.Equalf(t, expected, removeResult, "remove %d mismatch", i)
	}
	emptyLength := q.Length()
	require.Equalf(t, 0, emptyLength, "empty queue length mismatch")
}

//--

func TesLimitedPriorityHashQueueLength(t *testing.T) {
	testDefaultQueueLength(NewSimpleNothashableFactory(), NewLimitedPriorityHashQueue[SimpleNothashable](), t)
}

func TestPriorityLimitedPriorityHashQueueLength(t *testing.T) {
	testPriorityQueueLength(NewSimpleNothashableFactory(), NewPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedQueueNoLimitLength(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedQueueLength(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedPriorityQueueNoLimitLength(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedPriorityQueueLength(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TesHashLimitedPriorityHashQueueLength(t *testing.T) {
	testDefaultQueueLength(NewSimpleHashableFactory(), NewHashLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityHashLimitedPriorityHashQueueLength(t *testing.T) {
	testPriorityQueueLength(NewSimpleHashableFactory(), NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedQueueNoLimitLength(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedQueueLength(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedPriorityQueueNoLimitLength(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedPriorityQueueLength(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func testLimitedPriorityQueueNoLimitLength[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	testPriorityQueueLength(factory, func(priorityFun func(E) bool) Queue[E] {
		return makeLimitedPriorityQueueFun(priorityFun, 1500)
	}, t)
}

func testLimitedPriorityQueueLength[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	limit := 800
	q := makeLimitedPriorityQueueFun(priorityFunMod3[E], limit)
	testQueueLength(factory, q, 1000, limit, t)
}

func testLimitedQueueNoLimitLength[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	testDefaultQueueLength(factory, makeLimitedQueueFun(1500), t)
}

func testLimitedQueueLength[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	limit := 800
	q := makeLimitedQueueFun(limit)
	testQueueLength(factory, q, 1000, limit, t)
}

func testPriorityQueueLength[E IntBased](factory Factory[E], makePriorityQueueFun func(func(E) bool) Queue[E], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3[E])
	elementsToAdd := 1000
	testQueueLength(factory, q, elementsToAdd, elementsToAdd, t)
}

func testDefaultQueueLength[E IntBased](factory Factory[E], q Queue[E], t *testing.T) {
	elementsToAdd := 1000
	testQueueLength(factory, q, elementsToAdd, elementsToAdd, t)
}

func testQueueLength[E IntBased](factory Factory[E], q Queue[E], elementsToRemoveAdd, elementsToRemove int, t *testing.T) {
	emptyLength := q.Length()
	require.Equalf(t, 0, emptyLength, "empty queue length mismatch")

	for i := 0; i < elementsToRemoveAdd; i++ {
		require.Truef(t, q.Add(factory.Create(i)), "failed to add element %d", i)
		var expected int
		if i >= elementsToRemove {
			expected = elementsToRemove
		} else {
			expected = i + 1
		}
		currLength := q.Length()
		require.Equalf(t, expected, currLength, "adding %d: expected queue length mismatch", i)
	}
	for i := 0; i < elementsToRemove; i++ {
		q.Remove()
		currLength := q.Length()
		require.Equalf(t, elementsToRemove-i-1, currLength, "removing %d: expected queue length mismatch", i)
	}
}

//--

func TestLimitedPriorityHashQueueGet(t *testing.T) {
	testDefaultQueueGet(NewSimpleNothashableFactory(), NewLimitedPriorityHashQueue[SimpleNothashable](), t)
}

func TestPriorityLimitedPriorityHashQueueGet(t *testing.T) {
	testPriorityQueueGet(NewSimpleNothashableFactory(), NewPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedQueueNoLimitGet(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedQueueGet(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedPriorityQueueNoLimitGet(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedPriorityQueueGet(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestHashLimitedPriorityHashQueueGet(t *testing.T) {
	testDefaultQueueGet(NewSimpleHashableFactory(), NewHashLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityHashLimitedPriorityHashQueueGet(t *testing.T) {
	testPriorityQueueGet(NewSimpleHashableFactory(), NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedQueueNoLimitGet(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedQueueGet(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedPriorityQueueNoLimitGet(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedPriorityQueueGet(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func testLimitedPriorityQueueNoLimitGet[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	testPriorityQueueGet(factory, func(priorityFun func(E) bool) Queue[E] {
		return makeLimitedPriorityQueueFun(priorityFun, 1500)
	}, t)
}

func testLimitedPriorityQueueGet[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	limit := 800
	q := makeLimitedPriorityQueueFun(priorityFunMod2[E], limit)
	result := func(iteration int, index int) int {
		if index <= iteration/2 {
			return iteration - iteration%2 - 2*index
		}
		if iteration < limit {
			return -iteration + iteration%2 + 2*index - 1
		}
		return iteration + iteration%2 + 2*index - 2*limit + 1
	}
	testQueueGet(factory, q, 1000, result, t)
}

func testLimitedQueueNoLimitGet[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	testDefaultQueueGet(factory, makeLimitedQueueFun(1500), t)
}

func testLimitedQueueGet[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	limit := 800
	q := makeLimitedQueueFun(limit)
	result := func(iteration int, index int) int {
		if iteration < limit {
			return index
		}
		return index + iteration - limit + 1
	}
	testQueueGet(factory, q, 1000, result, t)
}

func testPriorityQueueGet[E IntBased](factory Factory[E], makePriorityQueueFun func(func(E) bool) Queue[E], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2[E])
	result := func(iteration int, index int) int {
		if index <= iteration/2 {
			return iteration - iteration%2 - 2*index
		}
		return -iteration + iteration%2 + 2*index - 1
	}
	testQueueGet(factory, q, 1000, result, t)
}

func testDefaultQueueGet[E IntBased](factory Factory[E], q Queue[E], t *testing.T) {
	testQueueGet(factory, q, 1000, func(iteration int, index int) int { return index }, t)
}

func testQueueGet[E IntBased](factory Factory[E], q Queue[E], elementsToAdd int, result func(iteration int, index int) int, t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Get test in short mode") // although it is not clear, why. Replacing require.Equalf in this code with `if a != b {t.Errorf(...)}` increases this test's performance significantly
	}
	for i := 0; i < elementsToAdd; i++ {
		require.Truef(t, q.Add(factory.Create(i)), "failed to add element %d", i)
		for j := 0; j < q.Length(); j++ {
			require.Equalf(t, factory.Create(result(i, j)), q.Get(j), "iteration %d index %d mismatch", i, j)
		}
	}
}

//--

func TestLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testDefaultQueueGetNegative(NewSimpleNothashableFactory(), NewLimitedPriorityHashQueue[SimpleNothashable](), t)
}

func TestPriorityLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testPriorityQueueGetNegative(NewSimpleNothashableFactory(), NewPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedQueueNoLimitGetNegative(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedQueueGetNegative(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedPriorityQueueNoLimitGetNegative(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedPriorityQueueGetNegative(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestHashLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testDefaultQueueGetNegative(NewSimpleHashableFactory(), NewHashLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityHashLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testPriorityQueueGetNegative(NewSimpleHashableFactory(), NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedQueueNoLimitGetNegative(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedQueueGetNegative(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedPriorityQueueNoLimitGetNegative(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedPriorityQueueGetNegative(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func testLimitedPriorityQueueNoLimitGetNegative[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	testPriorityQueueGetNegative(factory, func(priorityFun func(E) bool) Queue[E] {
		return makeLimitedPriorityQueueFun(priorityFun, 1500)
	}, t)
}

func testLimitedPriorityQueueGetNegative[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	limit := 800
	q := makeLimitedPriorityQueueFun(priorityFunMod2[E], limit)
	result := func(iteration int, index int) int {
		if iteration < limit {
			if index >= -(iteration+iteration%2)/2 {
				return iteration + iteration%2 + 2*index + 1
			}
			return -iteration - iteration%2 - 2*index - 2
		}
		if index <= (iteration-iteration%2)/2-limit {
			return iteration - iteration%2 - 2*index - 2*limit
		}
		return iteration + iteration%2 + 2*index + 1
	}
	testQueueGetNegative(factory, q, 1000, result, t)
}

func testLimitedQueueNoLimitGetNegative[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	testDefaultQueueGetNegative(factory, makeLimitedQueueFun(1500), t)
}

func testLimitedQueueGetNegative[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	testDefaultQueueGetNegative(factory, makeLimitedQueueFun(800), t)
}

func testPriorityQueueGetNegative[E IntBased](factory Factory[E], makePriorityQueueFun func(func(E) bool) Queue[E], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2[E])
	result := func(iteration int, index int) int {
		if index >= -(iteration+iteration%2)/2 {
			return iteration + iteration%2 + 2*index + 1
		}
		return -iteration - iteration%2 - 2*index - 2
	}
	testQueueGetNegative(factory, q, 1000, result, t)
}

func testDefaultQueueGetNegative[E IntBased](factory Factory[E], q Queue[E], t *testing.T) {
	testQueueGetNegative(factory, q, 1000, func(iteration int, index int) int { return iteration + index + 1 }, t)
}

func testQueueGetNegative[E IntBased](factory Factory[E], q Queue[E], elementsToAdd int, result func(iteration int, index int) int, t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GetNegative test in short mode") // although it is not clear, why. Replacing require.Equalf in this code with `if a != b {t.Errorf(...)}` increases this test's performance significantly
	}
	for i := 0; i < elementsToAdd; i++ {
		require.Truef(t, q.Add(factory.Create(i)), "failed to add element %d", i)
		for j := -1; j >= -q.Length(); j-- {
			require.Equalf(t, factory.Create(result(i, j)), q.Get(j), "iteration %d index %d mismatch", i, j)
		}
	}
}

//--

func TestLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testQueueGetOutOfRangePanics(NewSimpleNothashableFactory(), NewLimitedPriorityHashQueue[SimpleNothashable](), t)
}

func TestPriorityLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testPriorityQueueGetOutOfRangePanics(NewSimpleNothashableFactory(), NewPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedQueueGetOutOfRangePanics(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueGetOutOfRangePanics(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestHashLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testQueueGetOutOfRangePanics(NewSimpleHashableFactory(), NewHashLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityHashLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testPriorityQueueGetOutOfRangePanics(NewSimpleHashableFactory(), NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedQueueGetOutOfRangePanics(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueGetOutOfRangePanics(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func testLimitedPriorityQueueGetOutOfRangePanics[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	q := makeLimitedPriorityQueueFun(priorityFunMod2[E], 800)
	testQueueGetOutOfRangePanics(factory, q, t)
}

func testLimitedQueueGetOutOfRangePanics[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	testQueueGetOutOfRangePanics(factory, makeLimitedQueueFun(800), t)
}

func testPriorityQueueGetOutOfRangePanics[E IntBased](factory Factory[E], makePriorityQueueFun func(func(E) bool) Queue[E], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2[E])
	testQueueGetOutOfRangePanics(factory, q, t)
}

func testQueueGetOutOfRangePanics[E IntBased](factory Factory[E], q Queue[E], t *testing.T) {
	for i := 0; i < 3; i++ {
		require.Truef(t, q.Add(factory.Create(i)), "failed to add element %d", i)
	}
	require.Panicsf(t, func() { q.Get(-4) }, "should panic when too negative index")
	require.Panicsf(t, func() { q.Get(4) }, "should panic when index greater than length")
}

//--

func TestLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testQueuePeekOutOfRangePanics(NewSimpleNothashableFactory(), NewLimitedPriorityHashQueue[SimpleNothashable](), t)
}

func TestPriorityLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testPriorityQueuePeekOutOfRangePanics(NewSimpleNothashableFactory(), NewPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedQueuePeekOutOfRangePanics(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueuePeekOutOfRangePanics(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestHashtLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testQueuePeekOutOfRangePanics(NewSimpleHashableFactory(), NewHashLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityHashLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testPriorityQueuePeekOutOfRangePanics(NewSimpleHashableFactory(), NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedQueuePeekOutOfRangePanics(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueuePeekOutOfRangePanics(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func testLimitedPriorityQueuePeekOutOfRangePanics[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	q := makeLimitedPriorityQueueFun(priorityFunMod2[E], 800)
	testQueuePeekOutOfRangePanics(factory, q, t)
}

func testLimitedQueuePeekOutOfRangePanics[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	testQueuePeekOutOfRangePanics(factory, makeLimitedQueueFun(800), t)
}

func testPriorityQueuePeekOutOfRangePanics[E IntBased](factory Factory[E], makePriorityQueueFun func(func(E) bool) Queue[E], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2[E])
	testQueuePeekOutOfRangePanics(factory, q, t)
}

func testQueuePeekOutOfRangePanics[E IntBased](factory Factory[E], q Queue[E], t *testing.T) {
	require.Panicsf(t, func() { q.Peek() }, "should panic when peeking empty queue")
	require.Truef(t, q.Add(factory.Create(0)), "failed to add element 0")
	q.Remove()
	require.Panicsf(t, func() { q.Peek() }, "should panic when peeking emptied queue")
}

//--

func TestLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testQueueRemoveOutOfRangePanics(NewSimpleNothashableFactory(), NewLimitedPriorityHashQueue[SimpleNothashable](), t)
}

func TestPriorityLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testPriorityQueueRemoveOutOfRangePanics(NewSimpleNothashableFactory(), NewPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedQueueRemoveOutOfRangePanics(NewSimpleNothashableFactory(), NewLimitLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueRemoveOutOfRangePanics(NewSimpleNothashableFactory(), NewLimitPriorityLimitedPriorityHashQueue[SimpleNothashable], t)
}

func TestHashLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testQueueRemoveOutOfRangePanics(NewSimpleHashableFactory(), NewHashLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityHashLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testPriorityQueueRemoveOutOfRangePanics(NewSimpleHashableFactory(), NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedQueueRemoveOutOfRangePanics(NewSimpleHashableFactory(), NewLimitHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityHashLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueRemoveOutOfRangePanics(NewSimpleHashableFactory(), NewLimitPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func testLimitedPriorityQueueRemoveOutOfRangePanics[E IntBased](factory Factory[E], makeLimitedPriorityQueueFun func(priorityFun func(E) bool, limit int) Queue[E], t *testing.T) {
	q := makeLimitedPriorityQueueFun(priorityFunMod2[E], 800)
	testQueueRemoveOutOfRangePanics(factory, q, t)
}

func testLimitedQueueRemoveOutOfRangePanics[E IntBased](factory Factory[E], makeLimitedQueueFun func(limit int) Queue[E], t *testing.T) {
	testQueueRemoveOutOfRangePanics(factory, makeLimitedQueueFun(800), t)
}

func testPriorityQueueRemoveOutOfRangePanics[E IntBased](factory Factory[E], makePriorityQueueFun func(func(E) bool) Queue[E], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2[E])
	testQueueRemoveOutOfRangePanics(factory, q, t)
}

func testQueueRemoveOutOfRangePanics[E IntBased](factory Factory[E], q Queue[E], t *testing.T) {
	require.Panicsf(t, func() { q.Remove() }, "should panic when removing empty queue")
	require.Truef(t, q.Add(factory.Create(0)), "failed to add element 0")
	q.Remove()
	require.Panicsf(t, func() { q.Remove() }, "should panic when removing emptied queue")
}
