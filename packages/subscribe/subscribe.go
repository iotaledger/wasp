package subscribe

import (
	"fmt"
	"strings"
	"time"

	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/sub"
	_ "go.nanomsg.org/mangos/v3/transport/all"
)

func Subscribe(host string, messages chan<- []string, done <-chan bool, keepTrying bool, topics ...string) error {
	socket, err := sub.NewSocket()
	if err != nil {
		return err
	}
	for {
		err = socket.Dial("tcp://" + host)
		if err != nil {
			if keepTrying {
				time.Sleep(200 * time.Millisecond)
				continue
			} else {
				return fmt.Errorf("can't dial on sub socket %s: %s", host, err.Error())
			}
		}
		break
	}
	for _, topic := range topics {
		err = socket.SetOption(mangos.OptionSubscribe, []byte(topic))
	}
	if err != nil {
		return err
	}

	go func() {
		for {
			var buf []byte
			// fmt.Printf("recv\n")
			if buf, err = socket.Recv(); err != nil {
				close(messages)
				return
			}
			// fmt.Printf("received nanomsg '%s'\n", string(buf))
			if len(buf) > 0 {
				s := string(buf)
				messages <- strings.Split(s, " ")
			}
		}
	}()

	go func() {
		<-done
		socket.Close()
	}()

	return nil
}

type HostMessage struct {
	Sender  string
	Message []string
}

type Subscription struct {
	Hosts        []string
	Topics       []string
	HostMessages chan *HostMessage
	stopReading  chan bool
}

const (
	channelBufferSize  = 500
	channelLockTimeout = 1 * time.Second
)

//nolint:revive
func SubscribeMulti(hosts, topics []string, quorum ...int) (*Subscription, error) {
	if len(hosts) == 0 {
		return nil, fmt.Errorf("SubscribeMulti: no nanomsg hosts provided")
	}
	quorumNodes := len(hosts)
	if len(quorum) > 0 {
		if quorum[0] < 0 || quorum[0] >= len(hosts) {
			panic("invalid quorum value")
		}
		quorumNodes = quorum[0]
	}
	ret := &Subscription{
		Hosts:        hosts,
		Topics:       topics,
		HostMessages: make(chan *HostMessage, channelBufferSize),
		stopReading:  make(chan bool),
	}
	numSubscribed := 0
	for _, host := range hosts {
		hostMessages := make(chan []string)
		err := Subscribe(host, hostMessages, ret.stopReading, false, topics...)
		if err != nil {
			continue
		}
		numSubscribed++
		go func(host string) {
			for {
				select {
				case <-ret.stopReading:
					return
				case msg := <-hostMessages:
					select {
					case ret.HostMessages <- &HostMessage{host, msg}:
					case <-time.After(channelLockTimeout):
						// drop the host message if buffer of HostMessages is full
					}
				}
			}
		}(host)
	}
	if numSubscribed < quorumNodes {
		close(ret.stopReading)
		return nil, fmt.Errorf("SubscribeMulti: required %d nanomsg hosts, connected only to %d", quorumNodes, numSubscribed)
	}
	return ret, nil
}

func (subs *Subscription) WaitForPattern(pattern []string, timeout time.Duration, quorum ...int) bool {
	return subs.WaitForPatterns([][]string{pattern}, timeout, quorum...)
}

// WaitForPatterns waits until subscription receives all patterns from quorum of hosts
func (subs *Subscription) WaitForPatterns(patterns [][]string, timeout time.Duration, quorum ...int) bool {
	quorumNodes := len(subs.Hosts)
	if len(quorum) > 0 {
		if quorum[0] > 0 {
			quorumNodes = quorum[0]
		}
		if quorumNodes > len(subs.Hosts) {
			quorumNodes = len(subs.Hosts)
		}
	}
	received := make([]map[string]bool, len(patterns))
	for i := range received {
		received[i] = make(map[string]bool)
	}
	deadline := time.Now().Add(timeout)
	for {
		select {
		case m := <-subs.HostMessages:
			for i := range patterns {
				_, ok := received[i][m.Sender]
				if !ok {
					if matches(m.Message, patterns[i]) {
						received[i][m.Sender] = true
					}
				}
			}
			if checkQuorum(received, quorumNodes) {
				return true
			}

		case <-time.After(100 * time.Millisecond):
			if time.Now().After(deadline) {
				return false
			}
		}
	}
}

func checkQuorum(m []map[string]bool, quorum int) bool {
	for i := range m {
		if len(m[i]) < quorum {
			return false
		}
	}
	return true
}

func (subs *Subscription) Close() {
	close(subs.stopReading)
}

func matches(data, pattern []string) bool {
	size := len(pattern)
	if len(data) < size {
		size = len(data)
	}
	for i := 0; i < size; i++ {
		if pattern[i] == "*" {
			continue
		}
		if pattern[i] == data[i] {
			continue
		}
		return false
	}
	return true
}
