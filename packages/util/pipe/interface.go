package pipe

// minQueueLen is smallest capacity that queue may have.
const minQueueLen = 16

type Hashable interface {
	// GetHash should return a hash code value for the object. This method is supported
	// for the benefit of hash tables such as those provided by HashSet struct.
	// The general contract of GetHash is:
	// 	  * Whenever it is invoked on the same object more than once during an execution
	// 		of Wasp, the GetHash method must consistently return the same struct,
	//		provided no information used in Equals comparisons on the object is
	//		modified. This struct need not remain consistent from one execution
	//		of Wasp to another.
	//	  * If two objects are equal according to the Equals method, then calling
	//		the GetHash method on each of the two objects must produce the same struct.
	//	  * It is not required that if two objects are unequal according to the
	//		Equals method, then calling the GetHash method on each of the two objects
	//		must produce distinct struct. However, the programmer should be aware
	//		that producing distinct struct for unequal objects may improve the performance
	//		of hash tables.
	//	  *	the returned struct must be valid as a map key (e.g. no slices)
	GetHash() interface{}
	// Equals should indicate whether some other object is "equal to" this one.
	// The equals method implements an equivalence relation on non-null objects:
	//	  * It is reflexive: for any non-null value x, x.Equals(x) should return true.
	//	  * It is symmetric: for any non-null values x and y, x.Equals(y) should
	//		return true if and only if y.Equals(x) returns true.
	//	  * It is transitive: for any non-null values x, y, and z, if x.Equals(y)
	//		returns true and y.Equals(z) returns true, then x.Equals(z) should return true.
	//	  * It is consistent: for any non-null values x and y, multiple invocations
	//		of x.Equals(y) consistently return true or consistently return false,
	//		provided no information used in equals comparisons on the objects is modified.
	//    * For any non-null value x, x.Equals(null) should return false.
	Equals(elem interface{}) bool
}

type Set interface {
	Size() int
	Add(elem interface{}) bool
	Clear()
	Contains(elem interface{}) bool
	Remove(elem interface{}) bool
}

type Queue interface {
	Length() int
	Add(elem interface{}) bool
	Peek() interface{}
	Get(i int) interface{}
	Remove() interface{}
}

type Pipe interface {
	In() chan<- interface{}
	Out() <-chan interface{}
	Len() int
	Close()
}
