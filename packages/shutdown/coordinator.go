// Package shutdown provides functionality for coordinating the graceful shutdown
// of various components in the application.
package shutdown

import (
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"
)

// Coordinator implements a hierarchical shutdown await.
// This way the components can await for their sub-components before terminating.
type Coordinator struct {
	name       string // Unique name in the parent coordinator.
	path       string // For logging only.
	parent     *Coordinator
	nestedWG   *sync.WaitGroup
	nestedLock *sync.RWMutex
	nested     *shrinkingmap.ShrinkingMap[string, *Coordinator]
	log        log.Logger
}

func NewCoordinator(name string, log log.Logger) *Coordinator {
	return newCoordinator(name, nil, log)
}

func newCoordinator(name string, parent *Coordinator, log log.Logger) *Coordinator {
	path := name
	if parent != nil {
		path = parent.path + "." + name
	}
	return &Coordinator{
		name:       name,
		path:       path,
		parent:     parent,
		nestedWG:   &sync.WaitGroup{},
		nestedLock: &sync.RWMutex{},
		nested:     shrinkingmap.New[string, *Coordinator](),
		log:        log,
	}
}

// Nested makes a sub-context, with a name (for debugging).
func (s *Coordinator) Nested(name string) *Coordinator {
	s.nestedLock.Lock()
	defer s.nestedLock.Unlock()

	if nested, exists := s.nested.Get(name); exists {
		if !nested.CheckNestedDone() {
			panic(fmt.Errorf("nested context '%v' already exist at %v", name, s.path))
		}
	}
	newSub := newCoordinator(name, s, s.log)
	s.nested.Set(name, newSub)
	s.nestedWG.Add(1)
	return newSub
}

func (s *Coordinator) Done() {
	s.log.LogDebugf("context '%s' marked as done", s.path)
	if s.parent == nil {
		return
	}
	s.parent.subDone(s.name)
}

func (s *Coordinator) subDone(subName string) {
	s.nestedLock.Lock()
	defer s.nestedLock.Unlock()
	if !s.nested.Has(subName) {
		// Already marked as done.
		return
	}
	s.nested.Delete(subName)
	s.nestedWG.Done()
}

// WaitNestedWithLogging waits for all the hierarchy to complete. (same as Wait(), but logs what components are still being waited for)
func (s *Coordinator) WaitNestedWithLogging(logPeriod time.Duration) {
	nextLogTime := time.Now().Add(logPeriod)
	for {
		if s.CheckNestedDone() {
			return
		}
		now := time.Now()
		if now.After(nextLogTime) {
			s.nestedLock.RLock()
			s.nested.ForEachKey(func(name string) bool {
				s.log.LogDebugf("context '%s' waits for '%s' to complete", s.path, name)
				return true
			})
			s.nestedLock.RUnlock()
			nextLogTime = now.Add(logPeriod)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *Coordinator) WaitNested() {
	s.nestedWG.Wait()
}

// CheckNestedDone doesn't block, just checks if all the sub-tree is terminated.
func (s *Coordinator) CheckNestedDone() bool {
	s.nestedLock.RLock()
	defer s.nestedLock.RUnlock()
	return s.nested.Size() == 0
}
