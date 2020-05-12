package consensus

import (
	"bytes"
	"sort"
)

type sortByAge []*request

func (s sortByAge) Len() int {
	return len(s)
}

func (s sortByAge) Less(i, j int) bool {
	return s[i].whenMsgReceived.Before(s[j].whenMsgReceived)
}

func (s sortByAge) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortRequestsByAge(reqs []*request) {
	sort.Sort(sortByAge(reqs))
}

type sortById []*request

func (s sortById) Len() int {
	return len(s)
}

func (s sortById) Less(i, j int) bool {
	return bytes.Compare(s[i].reqId.Bytes(), s[j].reqId.Bytes()) < 0
}

func (s sortById) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortRequestsById(reqs []*request) {
	sort.Sort(sortById(reqs))
}
