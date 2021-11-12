package pipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testQueueBasicAddLengthPeekRemove(q Queue, elementsToAdd int, add func(index int) int, addResult func(index int) bool, elementsToRemove int, result func(index int) int, t *testing.T) {
	for i := 0; i < elementsToAdd; i++ {
		value := add(i)
		actualAddResult := q.Add(SimpleHashable(value))
		require.Equalf(t, addResult(i), actualAddResult, "add result of element %d value %d missmatch", i, value)
	}
	fullLength := q.Length()
	require.Equalf(t, elementsToRemove, fullLength, "full queue length missmatch")
	for i := 0; i < elementsToRemove; i++ {
		expected := SimpleHashable(result(i))
		peekResult := q.Peek().(SimpleHashable)
		require.Equalf(t, expected, peekResult, "peek %d missmatch", i)
		removeResult := q.Remove().(SimpleHashable)
		require.Equalf(t, expected, removeResult, "remove %d missmatch", i)
	}
	emptyLength := q.Length()
	require.Equalf(t, 0, emptyLength, "empty queue length missmatch")
}

//--

func TestDefaultLimitedPriorityHashQueueSimple(t *testing.T) {
	testDefaultQueueSimple(NewDefaultLimitedPriorityHashQueue(), t)
}

func TestPriorityLimitedPriorityHashQueueSimple(t *testing.T) {
	testPriorityQueueSimple(NewPriorityLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedQueueNoLimitSimple(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedQueueSimple(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitSimple(t *testing.T) {
	testLimitedPriorityQueueNoLimitSimple(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueSimple(t *testing.T) {
	testLimitedPriorityQueueSimple(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestHashLimitedPriorityHashQueueSimple(t *testing.T) {
	testDefaultQueueSimple(NewHashLimitedPriorityHashQueue(true), t)
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

func testLimitedPriorityQueueNoLimitSimple(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	testPriorityQueueSimple(func(priorityFun func(i interface{}) bool) Queue { return makeLimitedPriorityQueueFun(priorityFun, 15) }, t)
}

func testLimitedPriorityQueueSimple(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	resultArray := []int{9, 6, 3, 0, 4, 5, 7, 8}
	limit := len(resultArray)
	q := makeLimitedPriorityQueueFun(priorityFunMod3, limit)
	result := func(index int) int {
		return resultArray[index]
	}
	testQueueSimple(q, 10, limit, result, t)
}

func testLimitedQueueNoLimitSimple(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	testDefaultQueueSimple(makeLimitedQueueFun(15), t)
}

func testLimitedQueueSimple(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	limit := 8
	elementsToAdd := 10
	indexDiff := elementsToAdd - limit
	q := makeLimitedQueueFun(limit)
	result := func(index int) int {
		return index + indexDiff
	}
	testQueueSimple(q, elementsToAdd, limit, result, t)
}

func testPriorityQueueSimple(makePriorityQueueFun func(func(i interface{}) bool) Queue, t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3)
	resultArray := []int{9, 6, 3, 0, 1, 2, 4, 5, 7, 8}
	result := func(index int) int {
		return resultArray[index]
	}
	elementsToAdd := len(resultArray)
	testQueueSimple(q, elementsToAdd, elementsToAdd, result, t)
}

func testDefaultQueueSimple(q Queue, t *testing.T) {
	elementsToAdd := 10
	testQueueSimple(q, elementsToAdd, elementsToAdd, identityFunInt, t)
}

func testQueueSimple(q Queue, elementsToAdd, elementsToRemove int, result func(index int) int, t *testing.T) {
	testQueueBasicAddLengthPeekRemove(q, elementsToAdd, identityFunInt, alwaysTrueFun, elementsToRemove, result, t)
}

//--

func TestDefaultLimitedPriorityHashQueueTwice(t *testing.T) {
	testDefaultQueueTwice(NewDefaultLimitedPriorityHashQueue(), t)
}

func TestPriorityLimitedPriorityHashQueueTwice(t *testing.T) {
	testPriorityQueueTwice(NewPriorityLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testDefaultQueueTwice(NewLimitLimitedPriorityHashQueue(150), t)
}

func TestLimitLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 80
	elementsToAddSingle := 50
	indexDiff := 2*elementsToAddSingle - limit
	q := NewLimitLimitedPriorityHashQueue(limit)
	resultFun := func(index int) int {
		return (index + indexDiff) % elementsToAddSingle
	}
	testQueueTwice(q, elementsToAddSingle, alwaysTrueFun, limit, resultFun, t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testPriorityQueueTwice(func(priorityFun func(i interface{}) bool) Queue {
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
	testHashQueueTwice(NewHashLimitedPriorityHashQueue, t)
}

func TestPriorityHashLimitedPriorityHashQueueTwice(t *testing.T) {
	testPriorityHashQueueTwice(NewPriorityHashLimitedPriorityHashQueue, t)
}

func TestLimitHashLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testHashQueueTwice(func(hashNeeded bool) Queue { return NewLimitHashLimitedPriorityHashQueue(80, hashNeeded) }, t)
}

func TestLimitHashLimitedPriorityHashQueueTwice(t *testing.T) {
	limit := 30
	elementsToAddSingle := 50
	indexDiff := elementsToAddSingle - limit
	resultFun := func(index int) int { return index + indexDiff }
	q := NewLimitHashLimitedPriorityHashQueue(limit, true)
	testQueueTwice(q, elementsToAddSingle, alwaysTrueFun, limit, resultFun, t)
}

func TestLimitedPriorityHashQueueNoLimitTwice(t *testing.T) {
	testPriorityHashQueueTwice(func(priorityFun func(i interface{}) bool, hashNeeded bool) Queue {
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

func testHashQueueTwice(makeHashQueueFun func(hashNeeded bool) Queue, t *testing.T) {
	q := makeHashQueueFun(true)
	elementsToAddSingle := 50
	addResultFun := func(index int) bool { return index < elementsToAddSingle }
	testQueueTwice(q, elementsToAddSingle, addResultFun, elementsToAddSingle, identityFunInt, t)
}

func testPriorityHashQueueTwice(makePriorityHashQueueFun func(priorityFun func(i interface{}) bool, hashNeeded bool) Queue, t *testing.T) {
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

func testPriorityQueueTwice(makePriorityQueueFun func(func(i interface{}) bool) Queue, t *testing.T) {
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

func testDefaultQueueTwice(q Queue, t *testing.T) {
	elementsToAddSingle := 50
	resultFun := func(index int) int { return index % elementsToAddSingle }
	testQueueTwice(q, elementsToAddSingle, alwaysTrueFun, 2*elementsToAddSingle, resultFun, t)
}

func testQueueTwice(q Queue, elementsToAddSingle int, addResult func(index int) bool, elementsToRemove int, result func(index int) int, t *testing.T) {
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
	q := NewLimitPriorityLimitedPriorityHashQueue(func(i interface{}) bool {
		return i.(SimpleHashable) < cutOffSh
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
	q := NewLimitedPriorityHashQueue(func(i interface{}) bool {
		value := i.(SimpleHashable)
		return value < cutOffLow || cutOffHigh <= value
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
	testDefaultQueueAddRemove(NewDefaultLimitedPriorityHashQueue(), t)
}

func TestPriorityLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testPriorityQueueAddRemove(NewPriorityLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedQueueNoLimitAddRemove(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedQueueAddRemove(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitAddRemove(t *testing.T) {
	testLimitedPriorityQueueNoLimitAddRemove(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testLimitedPriorityQueueAddRemove(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestHashLimitedPriorityHashQueueAddRemove(t *testing.T) {
	testDefaultQueueAddRemove(NewHashLimitedPriorityHashQueue(true), t)
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

func testLimitedPriorityQueueNoLimitAddRemove(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	testPriorityQueueAddRemove(func(priorityFun func(i interface{}) bool) Queue { return makeLimitedPriorityQueueFun(priorityFun, 150) }, t)
}

func testLimitedPriorityQueueAddRemove(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
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

func testLimitedQueueNoLimitAddRemove(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	testDefaultQueueAddRemove(makeLimitedQueueFun(150), t)
}

func testLimitedQueueAddRemove(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
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

func testPriorityQueueAddRemove(makePriorityQueueFun func(func(i interface{}) bool) Queue, t *testing.T) {
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

func testDefaultQueueAddRemove(q Queue, t *testing.T) {
	elementsToAdd := 100
	elementsToRemoveAdd := 50
	testQueueAddRemove(q, elementsToAdd, elementsToRemoveAdd, elementsToAdd, func(index int) int { return index + elementsToRemoveAdd }, t)
}

func testQueueAddRemove(q Queue, elementsToAdd, elementsToRemoveAdd, elementsToRemove int, result func(index int) int, t *testing.T) {
	for i := 0; i < elementsToAdd; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
	}
	for i := 0; i < elementsToRemoveAdd; i++ {
		q.Remove()
		add := elementsToAdd + i
		require.Truef(t, q.Add(SimpleHashable(add)), "failed to add element %d", add)
	}
	fullLength := q.Length()
	require.Equalf(t, elementsToRemove, fullLength, "full queue length missmatch")

	for i := 0; i < elementsToRemove; i++ {
		expected := SimpleHashable(result(i))
		peekResult := q.Peek().(SimpleHashable)
		require.Equalf(t, expected, peekResult, "peek %d missmatch", i)
		removeResult := q.Remove().(SimpleHashable)
		require.Equalf(t, expected, removeResult, "remove %d missmatch", i)
	}
	emptyLength := q.Length()
	require.Equalf(t, 0, emptyLength, "empty queue length missmatch")
}

//--

func TesDefaultLimitedPriorityHashQueueLength(t *testing.T) {
	testDefaultQueueLength(NewDefaultLimitedPriorityHashQueue(), t)
}

func TestPriorityLimitedPriorityHashQueueLength(t *testing.T) {
	testPriorityQueueLength(NewPriorityLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedQueueNoLimitLength(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedQueueLength(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitLength(t *testing.T) {
	testLimitedPriorityQueueNoLimitLength(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueLength(t *testing.T) {
	testLimitedPriorityQueueLength(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TesHashLimitedPriorityHashQueueLength(t *testing.T) {
	testDefaultQueueLength(NewHashLimitedPriorityHashQueue(true), t)
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

func testLimitedPriorityQueueNoLimitLength(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	testPriorityQueueLength(func(priorityFun func(i interface{}) bool) Queue {
		return makeLimitedPriorityQueueFun(priorityFun, 1500)
	}, t)
}

func testLimitedPriorityQueueLength(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	limit := 800
	q := makeLimitedPriorityQueueFun(priorityFunMod3, limit)
	testQueueLength(q, 1000, limit, t)
}

func testLimitedQueueNoLimitLength(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	testDefaultQueueLength(makeLimitedQueueFun(1500), t)
}

func testLimitedQueueLength(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	limit := 800
	q := makeLimitedQueueFun(limit)
	testQueueLength(q, 1000, limit, t)
}

func testPriorityQueueLength(makePriorityQueueFun func(func(i interface{}) bool) Queue, t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod3)
	elementsToAdd := 1000
	testQueueLength(q, elementsToAdd, elementsToAdd, t)
}

func testDefaultQueueLength(q Queue, t *testing.T) {
	elementsToAdd := 1000
	testQueueLength(q, elementsToAdd, elementsToAdd, t)
}

func testQueueLength(q Queue, elementsToRemoveAdd, elementsToRemove int, t *testing.T) {
	emptyLength := q.Length()
	require.Equalf(t, 0, emptyLength, "empty queue length missmatch")

	for i := 0; i < elementsToRemoveAdd; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
		var expected int
		if i >= elementsToRemove {
			expected = elementsToRemove
		} else {
			expected = i + 1
		}
		currLength := q.Length()
		require.Equalf(t, expected, currLength, "adding %d: expected queue length missmatch", i)
	}
	for i := 0; i < elementsToRemove; i++ {
		q.Remove()
		currLength := q.Length()
		require.Equalf(t, elementsToRemove-i-1, currLength, "removing %d: expected queue length missmatch", i)
	}
}

//--

func TestDefaultLimitedPriorityHashQueueGet(t *testing.T) {
	testDefaultQueueGet(NewDefaultLimitedPriorityHashQueue(), t)
}

func TestPriorityLimitedPriorityHashQueueGet(t *testing.T) {
	testPriorityQueueGet(NewPriorityLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedQueueNoLimitGet(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedQueueGet(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitGet(t *testing.T) {
	testLimitedPriorityQueueNoLimitGet(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueGet(t *testing.T) {
	testLimitedPriorityQueueGet(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestHashLimitedPriorityHashQueueGet(t *testing.T) {
	testDefaultQueueGet(NewHashLimitedPriorityHashQueue(true), t)
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

func testLimitedPriorityQueueNoLimitGet(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	testPriorityQueueGet(func(priorityFun func(i interface{}) bool) Queue {
		return makeLimitedPriorityQueueFun(priorityFun, 1500)
	}, t)
}

func testLimitedPriorityQueueGet(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
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

func testLimitedQueueNoLimitGet(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	testDefaultQueueGet(makeLimitedQueueFun(1500), t)
}

func testLimitedQueueGet(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
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

func testPriorityQueueGet(makePriorityQueueFun func(func(i interface{}) bool) Queue, t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	result := func(iteration int, index int) int {
		if index <= iteration/2 {
			return iteration - iteration%2 - 2*index
		}
		return -iteration + iteration%2 + 2*index - 1
	}
	testQueueGet(q, 1000, result, t)
}

func testDefaultQueueGet(q Queue, t *testing.T) {
	testQueueGet(q, 1000, func(iteration int, index int) int { return index }, t)
}

func testQueueGet(q Queue, elementsToAdd int, result func(iteration int, index int) int, t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Get test in short mode") // although it is not clear, why. Replacing require.Equalf in this code with `if a != b {t.Errorf(...)}` increases this test's performance significantly
	}
	for i := 0; i < elementsToAdd; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
		for j := 0; j < q.Length(); j++ {
			require.Equalf(t, SimpleHashable(result(i, j)), q.Get(j).(SimpleHashable), "iteration %d index %d missmatch", i, j)
		}
	}
}

//--

func TestDefaultLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testDefaultQueueGetNegative(NewDefaultLimitedPriorityHashQueue(), t)
}

func TestPriorityLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testPriorityQueueGetNegative(NewPriorityLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedQueueNoLimitGetNegative(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedQueueGetNegative(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueNoLimitGetNegative(t *testing.T) {
	testLimitedPriorityQueueNoLimitGetNegative(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testLimitedPriorityQueueGetNegative(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestHashLimitedPriorityHashQueueGetNegative(t *testing.T) {
	testDefaultQueueGetNegative(NewHashLimitedPriorityHashQueue(true), t)
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

func testLimitedPriorityQueueNoLimitGetNegative(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	testPriorityQueueGetNegative(func(priorityFun func(i interface{}) bool) Queue {
		return makeLimitedPriorityQueueFun(priorityFun, 1500)
	}, t)
}

func testLimitedPriorityQueueGetNegative(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
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

func testLimitedQueueNoLimitGetNegative(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	testDefaultQueueGetNegative(makeLimitedQueueFun(1500), t)
}

func testLimitedQueueGetNegative(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	testDefaultQueueGetNegative(makeLimitedQueueFun(800), t)
}

func testPriorityQueueGetNegative(makePriorityQueueFun func(func(i interface{}) bool) Queue, t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	result := func(iteration int, index int) int {
		if index >= -(iteration+iteration%2)/2 {
			return iteration + iteration%2 + 2*index + 1
		}
		return -iteration - iteration%2 - 2*index - 2
	}
	testQueueGetNegative(q, 1000, result, t)
}

func testDefaultQueueGetNegative(q Queue, t *testing.T) {
	testQueueGetNegative(q, 1000, func(iteration int, index int) int { return iteration + index + 1 }, t)
}

func testQueueGetNegative(q Queue, elementsToAdd int, result func(iteration int, index int) int, t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GetNegative test in short mode") // although it is not clear, why. Replacing require.Equalf in this code with `if a != b {t.Errorf(...)}` increases this test's performance significantly
	}
	for i := 0; i < elementsToAdd; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
		for j := -1; j >= -q.Length(); j-- {
			require.Equalf(t, SimpleHashable(result(i, j)), q.Get(j).(SimpleHashable), "iteration %d index %d missmatch", i, j)
		}
	}
}

//--

func TestDefaultLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testQueueGetOutOfRangePanics(NewDefaultLimitedPriorityHashQueue(), t)
}

func TestPriorityLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testPriorityQueueGetOutOfRangePanics(NewPriorityLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedQueueGetOutOfRangePanics(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueGetOutOfRangePanics(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestHashLimitedPriorityHashQueueGetOutOfRangePanics(t *testing.T) {
	testQueueGetOutOfRangePanics(NewHashLimitedPriorityHashQueue(true), t)
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

func testLimitedPriorityQueueGetOutOfRangePanics(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	q := makeLimitedPriorityQueueFun(priorityFunMod2, 800)
	testQueueGetOutOfRangePanics(q, t)
}

func testLimitedQueueGetOutOfRangePanics(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	testQueueGetOutOfRangePanics(makeLimitedQueueFun(800), t)
}

func testPriorityQueueGetOutOfRangePanics(makePriorityQueueFun func(func(i interface{}) bool) Queue, t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	testQueueGetOutOfRangePanics(q, t)
}

func testQueueGetOutOfRangePanics(q Queue, t *testing.T) {
	for i := 0; i < 3; i++ {
		require.Truef(t, q.Add(SimpleHashable(i)), "failed to add element %d", i)
	}
	require.Panicsf(t, func() { q.Get(-4) }, "should panic when too negative index")
	require.Panicsf(t, func() { q.Get(4) }, "should panic when index greater than length")
}

//--

func TestDefaultLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testQueuePeekOutOfRangePanics(NewDefaultLimitedPriorityHashQueue(), t)
}

func TestPriorityLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testPriorityQueuePeekOutOfRangePanics(NewPriorityLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedQueuePeekOutOfRangePanics(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueuePeekOutOfRangePanics(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestHashtLimitedPriorityHashQueuePeekOutOfRangePanics(t *testing.T) {
	testQueuePeekOutOfRangePanics(NewHashLimitedPriorityHashQueue(true), t)
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

func testLimitedPriorityQueuePeekOutOfRangePanics(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	q := makeLimitedPriorityQueueFun(priorityFunMod2, 800)
	testQueuePeekOutOfRangePanics(q, t)
}

func testLimitedQueuePeekOutOfRangePanics(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	testQueuePeekOutOfRangePanics(makeLimitedQueueFun(800), t)
}

func testPriorityQueuePeekOutOfRangePanics(makePriorityQueueFun func(func(i interface{}) bool) Queue, t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	testQueuePeekOutOfRangePanics(q, t)
}

func testQueuePeekOutOfRangePanics(q Queue, t *testing.T) {
	require.Panicsf(t, func() { q.Peek() }, "should panic when peeking empty queue")
	require.Truef(t, q.Add(SimpleHashable(0)), "failed to add element 0")
	q.Remove()
	require.Panicsf(t, func() { q.Peek() }, "should panic when peeking emptied queue")
}

//--

func TestDefaultLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testQueueRemoveOutOfRangePanics(NewDefaultLimitedPriorityHashQueue(), t)
}

func TestPriorityLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testPriorityQueueRemoveOutOfRangePanics(NewPriorityLimitedPriorityHashQueue, t)
}

func TestLimitLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedQueueRemoveOutOfRangePanics(NewLimitLimitedPriorityHashQueue, t)
}

func TestLimitPriorityLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testLimitedPriorityQueueRemoveOutOfRangePanics(NewLimitPriorityLimitedPriorityHashQueue, t)
}

func TestHashLimitedPriorityHashQueueRemoveOutOfRangePanics(t *testing.T) {
	testQueueRemoveOutOfRangePanics(NewHashLimitedPriorityHashQueue(true), t)
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

func testLimitedPriorityQueueRemoveOutOfRangePanics(makeLimitedPriorityQueueFun func(priorityFun func(i interface{}) bool, limit int) Queue, t *testing.T) {
	q := makeLimitedPriorityQueueFun(priorityFunMod2, 800)
	testQueueRemoveOutOfRangePanics(q, t)
}

func testLimitedQueueRemoveOutOfRangePanics(makeLimitedQueueFun func(limit int) Queue, t *testing.T) {
	testQueueRemoveOutOfRangePanics(makeLimitedQueueFun(800), t)
}

func testPriorityQueueRemoveOutOfRangePanics(makePriorityQueueFun func(func(i interface{}) bool) Queue, t *testing.T) {
	q := makePriorityQueueFun(priorityFunMod2)
	testQueueRemoveOutOfRangePanics(q, t)
}

func testQueueRemoveOutOfRangePanics(q Queue, t *testing.T) {
	require.Panicsf(t, func() { q.Remove() }, "should panic when removing empty queue")
	require.Truef(t, q.Add(SimpleHashable(0)), "failed to add element 0")
	q.Remove()
	require.Panicsf(t, func() { q.Remove() }, "should panic when removing emptied queue")
}

//--

func newPriorityHashLimitedPriorityHashQueue(priorityFun func(i interface{}) bool) Queue {
	return NewPriorityHashLimitedPriorityHashQueue(priorityFun, true)
}

func newLimitHashLimitedPriorityHashQueue(limit int) Queue {
	return NewLimitHashLimitedPriorityHashQueue(limit, true)
}

func newLimitPriorityHashLimitedPriorityHashQueue(priorityFun func(i interface{}) bool, limit int) Queue {
	return NewLimitedPriorityHashQueue(priorityFun, limit, true)
}
