package pipe

type HashSet struct {
	hashTable map[interface{}][]Hashable
	count     int
}

var _ Set = &HashSet{}

// New constructs and returns a new Queue. Code duplication needed for benchmarks.
func NewHashSet() *HashSet {
	return &HashSet{
		hashTable: make(map[interface{}][]Hashable),
		count:     0,
	}
}

// Length returns the number of elements currently stored in the queue.
func (s *HashSet) Size() int {
	return s.count
}

func (s *HashSet) Add(elem interface{}) bool {
	hashable, ok := elem.(Hashable)
	if !ok {
		panic("Adding not hashable element to hash set")
	}
	hash := hashable.GetHash()
	hashEntry, ok := s.hashTable[hash]
	if ok {
		for i := 0; i < len(hashEntry); i++ {
			if hashEntry[i].Equals(hashable) {
				return false
			}
		}
		s.hashTable[hash] = append(hashEntry, hashable)
	} else {
		s.hashTable[hash] = []Hashable{hashable}
	}
	s.count++
	return true
}

func (s *HashSet) Clear() {
	s.hashTable = make(map[interface{}][]Hashable)
	s.count = 0
}

func (s *HashSet) Contains(elem interface{}) bool {
	hashable, ok := elem.(Hashable)
	if !ok {
		return false
	}
	hash := hashable.GetHash()
	hashEntry, ok := s.hashTable[hash]
	if !ok {
		return false
	}
	for i := 0; i < len(hashEntry); i++ {
		if hashEntry[i].Equals(hashable) {
			return true
		}
	}
	return false
}

func (s *HashSet) Remove(elem interface{}) bool {
	hashable, ok := elem.(Hashable)
	if !ok {
		return false
	}
	hash := hashable.GetHash()
	hashEntry, ok := s.hashTable[hash]
	if !ok {
		return false
	}
	length := len(hashEntry)
	for i := 0; i < length; i++ {
		if hashEntry[i].Equals(hashable) {
			if length == 1 {
				delete(s.hashTable, hash)
			} else {
				var newHashEntry []Hashable
				if i == length-1 {
					newHashEntry = hashEntry[0 : length-1]
				} else {
					newHashEntry = hashEntry[0:i]
					newHashEntry = append(newHashEntry, hashEntry[i+1:length]...)
				}
				s.hashTable[hash] = newHashEntry
			}
			s.count--
			return true
		}
	}
	return false
}
