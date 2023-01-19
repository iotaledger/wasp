package shutdowncoordinator

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
)

type ShutdownCoordinator struct {
	name          string
	parent        *ShutdownCoordinator
	wg            sync.WaitGroup
	subComponents map[string]bool
	log           *logger.Logger
}

func New(name string, parent *ShutdownCoordinator, log *logger.Logger) *ShutdownCoordinator {
	return &ShutdownCoordinator{
		name:          name,
		parent:        parent,
		wg:            sync.WaitGroup{},
		subComponents: make(map[string]bool),
		log:           log,
	}
}

// makes a subcontext, with a name (for debugging).
func (s *ShutdownCoordinator) Sub(name string) *ShutdownCoordinator {
	s.wg.Add(1)
	newSub := New(name, s, s.log)
	s.subComponents[name] = true
	return newSub
}

func (s *ShutdownCoordinator) Done() {
	if s.parent == nil {
		return
	}
	s.parent.subDone(s.name)
}

func (s *ShutdownCoordinator) subDone(subName string) {
	delete(s.subComponents, subName)
	s.wg.Done()
}

// waits to for all the hierarchy to complete. (same as Wait(), but logs what components are still being waited for)
func (s *ShutdownCoordinator) WaitWithLogging() {
	for {
		if s.AreAllSubComponentsDone() {
			return
		}
		for name := range s.subComponents {
			s.log.Debugf("%s waiting for %s to shutdown", s.name, name)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *ShutdownCoordinator) Wait() {
	s.wg.Wait()
}

// don't block, just check, if all the sub-tree is terminated.
func (s *ShutdownCoordinator) AreAllSubComponentsDone() bool {
	for name := range s.subComponents {
		s.log.Debugf("subcomponent %s is not done yet (parent: %s)", name, s.name)
	}
	return len(s.subComponents) == 0
}
