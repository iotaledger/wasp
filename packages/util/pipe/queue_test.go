package pipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testQueueBasicAddLengthPeekRemove(q Queue[SimpleHashable], elementsToAdd int, add func(index int) int, addResult func(index int) bool, elementsToRemove int, result func(index int) int, t *testing.T) {
	for i := 0; i < elementsToAdd; i++ {
		value := add(i)
		actualAddResult := q.Add(SimpleHashable(value))
		require.Equalf(t, addResult(i), actualAddResult, "add result of element %d value %d mismatch", i, value)
	}
	fullLength := q.Length()
	require.Equalf(t, elementsToRemove, fullLength, "full queue length mismatch")
	for i := 0; i < elementsToRemove; i++ {
		expected := SimpleHashable(result(i))
		peekResult := q.Peek()
		require.Equalf(t, expected, peekResult, "peek %d mismatch", i)
		removeResult := q.Remove()
		require.Equalf(t, expected, removeResult, "remove %d mismatch", i)
	}
	emptyLength := q.Length()
	require.Equalf(t, 0, emptyLength, "empty queue length mismatch")
}

//--

func TestDefaultLimitedPriorityHashQueueSimple(t *testing.T) {
	testDefaultQueueSimple(NewDefaultLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityLimitedPriorityHashQueueSimple(t *testing.T) {
	testPriorityQueueSimple(NewPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedQueueNoLimitSimple(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedQueueSimple(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedPriorityQueueNoLimitSimple(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedPriorityQueueSimple(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestHashLimitedPriorityHashQueueSimple(t *testing.T) {
	testDefaultQueueSimple(NewHashLimitedPriorityHashQueue[SimpleHashable](true), t)
}

func TestPriorityHashLimitedPriorityHashQueueSimple(t *testing.T) {
	testPriorityQueueSimple(newPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedQueueNoLimitSimple(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedQueueSimple(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedPriorityQueueNoLimitSimple(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedPriorityQueueSimple(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func testLimitedPriorityQueueNoLimitSimple(makeLimitedPriorityQueueFun func(priorityFun func(e SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	testPriorityQueueSimple(func(priorityFun func(e SimpleHashable) bool) Queue[SimpleHashable] {
		return makeLimitedPriorityQueueFun(priorityFun, 15)
	}, t)
}

func testLimitedPriorityQueueSimple(makeLimitedPriorityQueueFun func(priorityFun func(e SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	resultArray := []int{9, 6, 3, 0, 4, 5, 7, 8}
	limit := len(resultArray)
	q := makeLimitedPriorityQueueFun(priorityFunMod3, limit)
	result := func(index int) int {
		return resultArray[index]
	}
	testQueueSimple(q, 10, limit, result, t)
}

func testLimitedQueueNoLimitSimple(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	testDefaultQueueSimple(makeLimitedQueueFun(15), t)
}

func testLimitedQueueSimple(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	limit := 8
	elementsToAdd := 10
	indexDiff := elementsToAdd - limit
	q := makeLimitedQueueFun(limit)
	result := func(index int) int {
		return index + indexDiff
	}
	testQueueSimple(q, elementsToAdd, limit, result, t)
}

func testPriorityQueueSimple(makePriorityQueueFun func(func(e SimpleHashable) bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3)
	resultArray := []int{9, 6, 3, 0, 1, 2, 4, 5, 7, 8}
	result := func(index int) int {
		return resultArray[index]
	}
	elementsToAdd := len(resultArray)
	testQueueSimple(q, elementsToAdd, elementsToAdd, result, t)
}

func testDefaultQueueSimple(q Queue[SimpleHashable], t *testing.T) {
	elementsToAdd := 10
	testQueueSimple(q, elementsToAdd, elementsToAdd, identityFunInt, t)
}

func testQueueSimple(q Queue[SimpleHashable], elementsToAdd, elementsToRemove int, result func(index int) int, t *testing.T) {
	testQueueBasicAddLengthPeekRemove(q, elementsToAdd, identityFunInt, alwaysTrueFun, elementsToRemove, result, t)
}

//--

func TestDefaultLimitedPriorityHashQueueTwice(t *testing.T) {
	testDefaultQueueTwice(NewDefaultLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityLimitedPriorityHashQueueTwice(t *testing.T) {
	testPriorityQueueTwice(NewPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testDefaultQueueTwice(NewLimitLimitedPriorityHashQueue[SimpleHashable](150), t)
}

func TestLimitLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 80
	elementsToAddSingle := 50
	indexDiff := 2*elementsToAddSingle - limit
	q := NewLimitLimitedPriorityHashQueue[SimpleHashable](limit)
	resultFun := func(index int) int {
		return (index + indexDiff) % elementsToAddSingle
	}
	testQueueTwice(q, elementsToAddSingle, alwaysTrueFun, limit, resultFun, t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testPriorityQueueTwice(func(priorityFun func(i SimpleHashable) bool) Queue[SimpleHashable] {
		return NewLimitPriorityLimitedPriorityHashQueue(priorityFun, 150)
	}, t)
}

func TestLimitPriorityLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 80
	elementsToAddSingle := 50
	q := NewLimitPriorityLimitedPriorityHashQueue(priorityFunMod3, limit)
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
	testQueueTwice(q, elementsToAddSingle, alwaysTrueFun, limit, resultFun, t)
}

func TestHashLimitedPriorityHashQueueTwice(t *testing.T) {
	testHashQueueTwice(NewHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestPriorityHashLimitedPriorityHashQueueTwice(t *testing.T) {
	testPriorityHashQueueTwice(NewPriorityHashLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testHashQueueTwice(func(hashNeeded bool) Queue[SimpleHashable] {
		return NewLimitHashLimitedPriorityHashQueue[SimpleHashable](80, hashNeeded)
	}, t)
}

func TestLimitHashLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 30
	elementsToAddSingle := 50
	indexDiff := elementsToAddSingle - limit
	resultFun := func(index int) int { return index + indexDiff }
	q := NewLimitHashLimitedPriorityHashQueue[SimpleHashable](limit, true)
	testQueueTwice(q, elementsToAddSingle, alwaysTrueFun, limit, resultFun, t)
}

func TestLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testPriorityHashQueueTwice(func(priorityFun func(i SimpleHashable) bool, hashNeeded bool) Queue[SimpleHashable] {
		return NewLimitedPriorityHashQueue(priorityFun, 80, hashNeeded)
	}, t)
}

func TestLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 30
	elementsToAddSingle := 50
	q := NewLimitedPriorityHashQueue(priorityFunMod3, limit, true)
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
	testQueueTwice(q, elementsToAddSingle, addResultFun, limit, resultFun, t)
}

func testHashQueueTwice(makeHashQueueFun func(hashNeeded bool) Queue[SimpleHashable], t *testing.T) {
	q := makeHashQueueFun(true)
	elementsToAddSingle := 50
	addResultFun := func(index int) bool { return index < elementsToAddSingle }
	testQueueTwice(q, elementsToAddSingle, addResultFun, elementsToAddSingle, identityFunInt, t)
}

func testPriorityHashQueueTwice(makePriorityHashQueueFun func(priorityFun func(SimpleHashable) bool, hashNeeded bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityHashQueueFun(priorityFunMod3, true)
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
	testQueueTwice(q, elementsToAddSingle, addResultFun, elementsToAddSingle, resultFun, t)
}

func testPriorityQueueTwice(makePriorityQueueFun func(func(SimpleHashable) bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3)
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
	testQueueTwice(q, elementsToAddSingle, alwaysTrueFun, 2*elementsToAddSingle, resultFun, t)
}

func testDefaultQueueTwice(q Queue[SimpleHashable], t *testing.T) {
	elementsToAddSingle := 50
	resultFun := func(index int) int { return index % elementsToAddSingle }
	testQueueTwice(q, elementsToAddSingle, alwaysTrueFun, 2*elementsToAddSingle, resultFun, t)
}

func testQueueTwice(q Queue[SimpleHashable], elementsToAddSingle int, addResult func(index int) bool, elementsToRemove int, result func(index int) int, t *testing.T) {
	addFun := func(index int) int {
		return index % elementsToAddSingle
	}
	testQueueBasicAddLengthPeekRemove(q, 2*elementsToAddSingle, addFun, addResult, elementsToRemove, result, t)
}

//--

func TestLimitPriorityLimitedPriorityHashQueueOverflow(t *testing.T) {
	limit := 30
	elementsToAddSingle := 50
	cutOff := elementsToAddSingle / 2
	cutOffSh := SimpleHashable(cutOff)
	q := NewLimitPriorityLimitedPriorityHashQueue(func(e SimpleHashable) bool {
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
	testQueueTwice(q, elementsToAddSingle, addResultFun, limit, resultFun, t)
}

func TestLimitedPriorityHashQueueOverflow(t *testing.T) {
	limit := 30
	elementsToAddSingle := 50
	cutOffLow := SimpleHashable(20)
	cutOffHigh := SimpleHashable(40)
	q := NewLimitedPriorityHashQueue(func(e SimpleHashable) bool {
		return e < cutOffLow || cutOffHigh <= e
	}, limit, true)
	addResultFun := func(index int) bool {
		return index < elementsToAddSingle
	}
	resultFun := func(index int) int {
		if index < 10 {
			return 49 - index
		}
		return 29 - index
	}
	testQueueTwice(q, elementsToAddSingle, addResultFun, limit, resultFun, t)
}

//--

func TestLimitedPriorityHashQueueDuplicates(t *testing.T) {
	limit := 80
	elementsToAddFirstIteration := 50
	q := NewLimitedPriorityHashQueue(priorityFunMod3, limit, true)
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
	testQueueBasicAddLengthPeekRemove(q, 3*elementsToAddFirstIteration, addFun, addResultFun, limit, resultFun, t)
}

//--

func TestDefaultLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testDefaultQueueAddRemove(NewDefaultLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testPriorityQueueAddRemove(NewPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedQueueNoLimitAddRemove(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedQueueAddRemove(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedPriorityQueueNoLimitAddRemove(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedPriorityQueueAddRemove(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestHashLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testDefaultQueueAddRemove(NewHashLimitedPriorityHashQueue[SimpleHashable](true), t)
}

func TestPriorityHashLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testPriorityQueueAddRemove(newPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedQueueNoLimitAddRemove(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedQueueAddRemove(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedPriorityQueueNoLimitAddRemove(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedPriorityQueueAddRemove(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func testLimitedPriorityQueueNoLimitAddRemove(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	testPriorityQueueAddRemove(func(priorityFun func(SimpleHashable) bool) Queue[SimpleHashable] {
		return makeLimitedPriorityQueueFun(priorityFun, 150)
	}, t)
}

func testLimitedPriorityQueueAddRemove(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	limit := 80
	q := makeLimitedPriorityQueueFun(priorityFunMod3, limit)
	result := func(index int) int {
		if index%2 == 0 {
			return 3*index/2 + 31
		}
		return (3*index + 61) / 2
	}
	testQueueAddRemove(q, 100, 50, limit, result, t)
}

func testLimitedQueueNoLimitAddRemove(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	testDefaultQueueAddRemove(makeLimitedQueueFun(150), t)
}

func testLimitedQueueAddRemove(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	limit := 80
	elementsToAdd := 100
	elementsToRemoveAdd := 50
	indexDiff := elementsToAdd - limit + elementsToRemoveAdd
	q := makeLimitedQueueFun(limit)
	result := func(index int) int {
		return index + indexDiff
	}
	testQueueAddRemove(q, elementsToAdd, elementsToRemoveAdd, limit, result, t)
}

func testPriorityQueueAddRemove(makePriorityQueueFun func(func(SimpleHashable) bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3)
	result := func(index int) int {
		if index%2 == 0 {
			return 3*index/2 + 1
		}
		return (3*index + 1) / 2
	}
	elementsToAdd := 100
	testQueueAddRemove(q, elementsToAdd, 50, elementsToAdd, result, t)
}

func testDefaultQueueAddRemove(q Queue[SimpleHashable], t *testing.T) {
	elementsToAdd := 100
	elementsToRemoveAdd := 50
	testQueueAddRemove(q, elementsToAdd, elementsToRemoveAdd, elementsToAdd, func(index int) int { return index + elementsToRemoveAdd }, t)
}

func testQueueAddRemove(q Queue[SimpleHashable], elementsToAdd, elementsToRemoveAdd, elementsToRemove int, result func(index int) int, t *testing.T) {
	for i := 0; i < elementsToAdd; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
	}
	for i := 0; i < elementsToRemoveAdd; i++ {
		q.Remove()
		add := elementsToAdd + i
		require.Truef(t, q.Add(SimpleHashable(add)), "failed to add element %d", add)
	}
	fullLength := q.Length()
	require.Equalf(t, elementsToRemove, fullLength, "full queue length mismatch")

	for i := 0; i < elementsToRemove; i++ {
		expected := SimpleHashable(result(i))
		peekResult := q.Peek()
		require.Equalf(t, expected, peekResult, "peek %d mismatch", i)
		removeResult := q.Remove()
		require.Equalf(t, expected, removeResult, "remove %d mismatch", i)
	}
	emptyLength := q.Length()
	require.Equalf(t, 0, emptyLength, "empty queue length mismatch")
}

//--

func TesDefaultLimitedPriorityHashQueueLength(t *testing.T) {
	testDefaultQueueLength(NewDefaultLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityLimitedPriorityHashQueueLength(t *testing.T) {
	testPriorityQueueLength(NewPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedQueueNoLimitLength(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedQueueLength(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedPriorityQueueNoLimitLength(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedPriorityQueueLength(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TesHashLimitedPriorityHashQueueLength(t *testing.T) {
	testDefaultQueueLength(NewHashLimitedPriorityHashQueue[SimpleHashable](true), t)
}

func TestPriorityHashLimitedPriorityHashQueueLength(t *testing.T) {
	testPriorityQueueLength(newPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedQueueNoLimitLength(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedQueueLength(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedPriorityQueueNoLimitLength(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedPriorityQueueLength(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func testLimitedPriorityQueueNoLimitLength(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	testPriorityQueueLength(func(priorityFun func(SimpleHashable) bool) Queue[SimpleHashable] {
		return makeLimitedPriorityQueueFun(priorityFun, 1500)
	}, t)
}

func testLimitedPriorityQueueLength(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	limit := 800
	q := makeLimitedPriorityQueueFun(priorityFunMod3, limit)
	testQueueLength(q, 1000, limit, t)
}

func testLimitedQueueNoLimitLength(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	testDefaultQueueLength(makeLimitedQueueFun(1500), t)
}

func testLimitedQueueLength(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	limit := 800
	q := makeLimitedQueueFun(limit)
	testQueueLength(q, 1000, limit, t)
}

func testPriorityQueueLength(makePriorityQueueFun func(func(SimpleHashable) bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3)
	elementsToAdd := 1000
	testQueueLength(q, elementsToAdd, elementsToAdd, t)
}

func testDefaultQueueLength(q Queue[SimpleHashable], t *testing.T) {
	elementsToAdd := 1000
	testQueueLength(q, elementsToAdd, elementsToAdd, t)
}

func testQueueLength(q Queue[SimpleHashable], elementsToRemoveAdd, elementsToRemove int, t *testing.T) {
	emptyLength := q.Length()
	require.Equalf(t, 0, emptyLength, "empty queue length mismatch")

	for i := 0; i < elementsToRemoveAdd; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
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

func TestDefaultLimitedPriorityHashQueueGet(t *testing.T) {
	testDefaultQueueGet(NewDefaultLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityLimitedPriorityHashQueueGet(t *testing.T) {
	testPriorityQueueGet(NewPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedQueueNoLimitGet(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedQueueGet(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedPriorityQueueNoLimitGet(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedPriorityQueueGet(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestHashLimitedPriorityHashQueueGet(t *testing.T) {
	testDefaultQueueGet(NewHashLimitedPriorityHashQueue[SimpleHashable](true), t)
}

func TestPriorityHashLimitedPriorityHashQueueGet(t *testing.T) {
	testPriorityQueueGet(newPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedQueueNoLimitGet(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedQueueGet(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedPriorityQueueNoLimitGet(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedPriorityQueueGet(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func testLimitedPriorityQueueNoLimitGet(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	testPriorityQueueGet(func(priorityFun func(SimpleHashable) bool) Queue[SimpleHashable] {
		return makeLimitedPriorityQueueFun(priorityFun, 1500)
	}, t)
}

func testLimitedPriorityQueueGet(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	limit := 800
	q := makeLimitedPriorityQueueFun(priorityFunMod2, limit)
	result := func(iteration int, index int) int {
		if index <= iteration/2 {
			return iteration - iteration%2 - 2*index
		}
		if iteration < limit {
			return -iteration + iteration%2 + 2*index - 1
		}
		return iteration + iteration%2 + 2*index - 2*limit + 1
	}
	testQueueGet(q, 1000, result, t)
}

func testLimitedQueueNoLimitGet(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	testDefaultQueueGet(makeLimitedQueueFun(1500), t)
}

func testLimitedQueueGet(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	limit := 800
	q := makeLimitedQueueFun(limit)
	result := func(iteration int, index int) int {
		if iteration < limit {
			return index
		}
		return index + iteration - limit + 1
	}
	testQueueGet(q, 1000, result, t)
}

func testPriorityQueueGet(makePriorityQueueFun func(func(SimpleHashable) bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	result := func(iteration int, index int) int {
		if index <= iteration/2 {
			return iteration - iteration%2 - 2*index
		}
		return -iteration + iteration%2 + 2*index - 1
	}
	testQueueGet(q, 1000, result, t)
}

func testDefaultQueueGet(q Queue[SimpleHashable], t *testing.T) {
	testQueueGet(q, 1000, func(iteration int, index int) int { return index }, t)
}

func testQueueGet(q Queue[SimpleHashable], elementsToAdd int, result func(iteration int, index int) int, t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Get test in short mode") // although it is not clear, why. Replacing require.Equalf in this code with `if a != b {t.Errorf(...)}` increases this test's performance significantly
	}
	for i := 0; i < elementsToAdd; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
		for j := 0; j < q.Length(); j++ {
			require.Equalf(t, SimpleHashable(result(i, j)), q.Get(j), "iteration %d index %d mismatch", i, j)
		}
	}
}

//--

func TestDefaultLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testDefaultQueueGetNegative(NewDefaultLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testPriorityQueueGetNegative(NewPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedQueueNoLimitGetNegative(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedQueueGetNegative(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedPriorityQueueNoLimitGetNegative(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedPriorityQueueGetNegative(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestHashLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testDefaultQueueGetNegative(NewHashLimitedPriorityHashQueue[SimpleHashable](true), t)
}

func TestPriorityHashLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testPriorityQueueGetNegative(newPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedQueueNoLimitGetNegative(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedQueueGetNegative(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedPriorityQueueNoLimitGetNegative(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedPriorityQueueGetNegative(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func testLimitedPriorityQueueNoLimitGetNegative(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	testPriorityQueueGetNegative(func(priorityFun func(SimpleHashable) bool) Queue[SimpleHashable] {
		return makeLimitedPriorityQueueFun(priorityFun, 1500)
	}, t)
}

func testLimitedPriorityQueueGetNegative(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	limit := 800
	q := makeLimitedPriorityQueueFun(priorityFunMod2, limit)
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
	testQueueGetNegative(q, 1000, result, t)
}

func testLimitedQueueNoLimitGetNegative(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	testDefaultQueueGetNegative(makeLimitedQueueFun(1500), t)
}

func testLimitedQueueGetNegative(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	testDefaultQueueGetNegative(makeLimitedQueueFun(800), t)
}

func testPriorityQueueGetNegative(makePriorityQueueFun func(func(SimpleHashable) bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	result := func(iteration int, index int) int {
		if index >= -(iteration+iteration%2)/2 {
			return iteration + iteration%2 + 2*index + 1
		}
		return -iteration - iteration%2 - 2*index - 2
	}
	testQueueGetNegative(q, 1000, result, t)
}

func testDefaultQueueGetNegative(q Queue[SimpleHashable], t *testing.T) {
	testQueueGetNegative(q, 1000, func(iteration int, index int) int { return iteration + index + 1 }, t)
}

func testQueueGetNegative(q Queue[SimpleHashable], elementsToAdd int, result func(iteration int, index int) int, t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GetNegative test in short mode") // although it is not clear, why. Replacing require.Equalf in this code with `if a != b {t.Errorf(...)}` increases this test's performance significantly
	}
	for i := 0; i < elementsToAdd; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
		for j := -1; j >= -q.Length(); j-- {
			require.Equalf(t, SimpleHashable(result(i, j)), q.Get(j), "iteration %d index %d mismatch", i, j)
		}
	}
}

//--

func TestDefaultLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testQueueGetOutOfRangePanics(NewDefaultLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testPriorityQueueGetOutOfRangePanics(NewPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedQueueGetOutOfRangePanics(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueGetOutOfRangePanics(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestHashLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testQueueGetOutOfRangePanics(NewHashLimitedPriorityHashQueue[SimpleHashable](true), t)
}

func TestPriorityHashLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testPriorityQueueGetOutOfRangePanics(newPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedQueueGetOutOfRangePanics(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueGetOutOfRangePanics(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func testLimitedPriorityQueueGetOutOfRangePanics(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	q := makeLimitedPriorityQueueFun(priorityFunMod2, 800)
	testQueueGetOutOfRangePanics(q, t)
}

func testLimitedQueueGetOutOfRangePanics(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	testQueueGetOutOfRangePanics(makeLimitedQueueFun(800), t)
}

func testPriorityQueueGetOutOfRangePanics(makePriorityQueueFun func(func(SimpleHashable) bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	testQueueGetOutOfRangePanics(q, t)
}

func testQueueGetOutOfRangePanics(q Queue[SimpleHashable], t *testing.T) {
	for i := 0; i < 3; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
	}
	require.Panicsf(t, func() { q.Get(-4) }, "should panic when too negative index")
	require.Panicsf(t, func() { q.Get(4) }, "should panic when index greater than length")
}

//--

func TestDefaultLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testQueuePeekOutOfRangePanics(NewDefaultLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testPriorityQueuePeekOutOfRangePanics(NewPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedQueuePeekOutOfRangePanics(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueuePeekOutOfRangePanics(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestHashtLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testQueuePeekOutOfRangePanics(NewHashLimitedPriorityHashQueue[SimpleHashable](true), t)
}

func TestPriorityHashLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testPriorityQueuePeekOutOfRangePanics(newPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedQueuePeekOutOfRangePanics(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueuePeekOutOfRangePanics(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func testLimitedPriorityQueuePeekOutOfRangePanics(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	q := makeLimitedPriorityQueueFun(priorityFunMod2, 800)
	testQueuePeekOutOfRangePanics(q, t)
}

func testLimitedQueuePeekOutOfRangePanics(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	testQueuePeekOutOfRangePanics(makeLimitedQueueFun(800), t)
}

func testPriorityQueuePeekOutOfRangePanics(makePriorityQueueFun func(func(SimpleHashable) bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	testQueuePeekOutOfRangePanics(q, t)
}

func testQueuePeekOutOfRangePanics(q Queue[SimpleHashable], t *testing.T) {
	require.Panicsf(t, func() { q.Peek() }, "should panic when peeking empty queue")
	require.Truef(t, q.Add(SimpleHashable(0)), "failed to add element 0")
	q.Remove()
	require.Panicsf(t, func() { q.Peek() }, "should panic when peeking emptied queue")
}

//--

func TestDefaultLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testQueueRemoveOutOfRangePanics(NewDefaultLimitedPriorityHashQueue[SimpleHashable](), t)
}

func TestPriorityLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testPriorityQueueRemoveOutOfRangePanics(NewPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedQueueRemoveOutOfRangePanics(NewLimitLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestLimitPriorityLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueRemoveOutOfRangePanics(NewLimitPriorityLimitedPriorityHashQueue[SimpleHashable], t)
}

func TestHashLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testQueueRemoveOutOfRangePanics(NewHashLimitedPriorityHashQueue[SimpleHashable](true), t)
}

func TestPriorityHashLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testPriorityQueueRemoveOutOfRangePanics(newPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedQueueRemoveOutOfRangePanics(newLimitHashLimitedPriorityHashQueue, t)
}

func TestLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueRemoveOutOfRangePanics(newLimitPriorityHashLimitedPriorityHashQueue, t)
}

func testLimitedPriorityQueueRemoveOutOfRangePanics(makeLimitedPriorityQueueFun func(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable], t *testing.T) {
	q := makeLimitedPriorityQueueFun(priorityFunMod2, 800)
	testQueueRemoveOutOfRangePanics(q, t)
}

func testLimitedQueueRemoveOutOfRangePanics(makeLimitedQueueFun func(limit int) Queue[SimpleHashable], t *testing.T) {
	testQueueRemoveOutOfRangePanics(makeLimitedQueueFun(800), t)

}
func testPriorityQueueRemoveOutOfRangePanics(makePriorityQueueFun func(func(SimpleHashable) bool) Queue[SimpleHashable], t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	testQueueRemoveOutOfRangePanics(q, t)
}

func testQueueRemoveOutOfRangePanics(q Queue[SimpleHashable], t *testing.T) {
	require.Panicsf(t, func() { q.Remove() }, "should panic when removing empty queue")
	require.Truef(t, q.Add(SimpleHashable(0)), "failed to add element 0")
	q.Remove()
	require.Panicsf(t, func() { q.Remove() }, "should panic when removing emptied queue")
}

//--

func newPriorityHashLimitedPriorityHashQueue(priorityFun func(SimpleHashable) bool) Queue[SimpleHashable] {
	return NewPriorityHashLimitedPriorityHashQueue(priorityFun, true)
}

func newLimitHashLimitedPriorityHashQueue(limit int) Queue[SimpleHashable] {
	return NewLimitHashLimitedPriorityHashQueue[SimpleHashable](limit, true)
}

func newLimitPriorityHashLimitedPriorityHashQueue(priorityFun func(SimpleHashable) bool, limit int) Queue[SimpleHashable] {
	return NewLimitedPriorityHashQueue[SimpleHashable](priorityFun, limit, true)
}
