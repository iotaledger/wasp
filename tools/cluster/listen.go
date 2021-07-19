package cluster

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/iotaledger/wasp/packages/subscribe"
)

type MessageCounter struct {
	cluster      *Cluster
	subscription *subscribe.Subscription
	expectations map[string]int
	counters     map[string]map[string]int
}

func NewMessageCounter(cluster *Cluster, nodes []int, expectations map[string]int) (*MessageCounter, error) {
	m := &MessageCounter{
		cluster:      cluster,
		expectations: expectations,
		counters:     make(map[string]map[string]int),
	}

	topics := make([]string, 0)
	for t := range expectations {
		topics = append(topics, t)
	}
	sort.Strings(topics)

	allNodesNanomsg := cluster.Config.NanomsgHosts(nodes)
	for _, host := range allNodesNanomsg {
		m.counters[host] = make(map[string]int)
		for msgType := range expectations {
			m.counters[host][msgType] = 0
		}
	}

	var err error
	m.subscription, err = subscribe.SubscribeMulti(allNodesNanomsg, topics)
	return m, err
}

func (m *MessageCounter) CollectMessages(duration time.Duration) {
	fmt.Printf("[cluster] collecting publisher's messages for %v\n", duration)

	deadline := time.Now().Add(duration)
	for {
		select {
		case msg := <-m.subscription.HostMessages:
			m.countMessage(msg)

		case <-time.After(500 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			break
		}
	}
}

func (m *MessageCounter) WaitUntilExpectationsMet() bool {
	fmt.Printf("[cluster] collecting publisher's messages\n")

	for {
		fail, pass, report := m.report()
		if fail {
			fmt.Printf("\n[cluster] Message expectations failed for '%s':\n%s\n", m.cluster.Name, report)
			return false
		}
		if pass {
			return true
		}

		select {
		case msg := <-m.subscription.HostMessages:
			m.countMessage(msg)
		case <-time.After(90 * time.Second):
			return m.Report()
		}
	}
}

func (m *MessageCounter) countMessage(msg *subscribe.HostMessage) {
	m.counters[msg.Sender][msg.Message[0]]++
}

func (m *MessageCounter) Report() bool {
	_, pass, report := m.report()
	fmt.Printf("\n[cluster] Message statistics for '%s':\n%s\n", m.cluster.Name, report)
	return pass
}

func (m *MessageCounter) report() (bool, bool, string) {
	fail := false
	pass := true
	report := ""
	for host, counters := range m.counters {
		report += fmt.Sprintf("Node: %s\n", host)
		for _, t := range m.subscription.Topics {
			res := counters[t]
			exp := m.expectations[t]
			e := "-"
			f := ""
			if exp >= 0 {
				e = strconv.Itoa(exp)
				if res == exp {
					f = "ok"
				} else {
					f = "fail"
					pass = false
					if res > exp {
						// got more messages than expected, no need to keep running
						fail = true
					}
				}
			}
			report += fmt.Sprintf("          %s: %d (%s) %s\n", t, res, e, f)
		}
	}
	return fail, pass, report
}

func (m *MessageCounter) Close() {
	m.subscription.Close()
}
