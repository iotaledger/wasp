package subscribe

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"sync/atomic"
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
			//fmt.Printf("recv\n")
			if buf, err = socket.Recv(); err != nil {
				close(messages)
				return
			}
			//fmt.Printf("received nanomsg '%s'\n", string(buf))
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

func SubscribeMultiOld(hosts []string, messages chan<- *HostMessage, done chan bool, topics ...string) error {
	for _, host := range hosts {
		hostMessages := make(chan []string)
		err := Subscribe(host, hostMessages, done, false, topics...)
		if err != nil {
			return err
		}
		go func(host string) {
			for {
				select {
				case <-done:
					return
				case msg := <-hostMessages:
					messages <- &HostMessage{host, msg}
				}
			}
		}(host)
	}
	return nil
}

type Subscription struct {
	chHostMessages chan *HostMessage
	chStopReading  chan bool
	hosts          []string
	numSubscribed  int
}

const (
	channelBufferSize  = 500
	channelLockTimeout = 1 * time.Second
)

func SubscribeMulti(hosts []string, topics string, quorum ...int) (*Subscription, error) {
	if len(hosts) == 0 {
		return nil, fmt.Errorf("SubscribeMulti: no nanomsg hosts provided")
	}
	quorumNodes := len(hosts)
	if len(quorum) > 0 {
		if quorum[0] > 0 || quorum[0] < len(hosts) {
			quorumNodes = quorum[0]
		}
	}
	ret := &Subscription{
		chHostMessages: make(chan *HostMessage, channelBufferSize),
		chStopReading:  make(chan bool),
		hosts:          hosts,
	}
	for _, host := range hosts {
		hostMessages := make(chan []string)
		err := Subscribe(host, hostMessages, ret.chStopReading, false, []string{topics}...)
		if err != nil {
			continue
		}
		ret.numSubscribed++
		go func(host string) {
			for {
				select {
				case <-ret.chStopReading:
					return
				case msg := <-hostMessages:
					ret.chHostMessages <- &HostMessage{host, msg}
				case <-time.After(channelLockTimeout):
					// loosing the host message if buffer of chHostMessages is full
				}
			}
		}(host)
	}
	if ret.numSubscribed < quorumNodes {
		close(ret.chStopReading)
		return nil, fmt.Errorf("SubscribeMulti: required %d nanomsg hosts, connected only to %d", quorumNodes, ret.numSubscribed)
	}
	return ret, nil
}

func (subs *Subscription) WaitForPattern(pattern []string, timeout time.Duration, quorum ...int) bool {
	return subs.WaitForPatterns([][]string{pattern}, timeout, quorum...)
}

// WaitForPatterns waits until subscription receives all patterns from quorum of hosts
func (subs *Subscription) WaitForPatterns(patterns [][]string, timeout time.Duration, quorum ...int) bool {
	quorumNodes := len(subs.hosts)
	if len(quorum) > 0 {
		if quorum[0] > 0 {
			quorumNodes = quorum[0]
		}
		if quorumNodes > len(subs.hosts) {
			quorumNodes = len(subs.hosts)
		}
	}
	received := make([]map[string]bool, len(patterns))
	for i := range received {
		received[i] = make(map[string]bool)
	}
	deadline := time.Now().Add(timeout)
	for {
		select {
		case m := <-subs.chHostMessages:
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
	close(subs.chStopReading)
}

func subscribeAndWaitForPattern(host string, pattern []string, timeout time.Duration) (bool, error) {
	if len(pattern) == 0 {
		return false, fmt.Errorf("wrong pattern")
	}
	socket, err := sub.NewSocket()
	if err != nil {
		return false, err
	}
	err = socket.Dial("tcp://" + host)
	if err != nil {
		return false, fmt.Errorf("can't dial on sub socket %s: %s", host, err.Error())
	}
	err = socket.SetOption(mangos.OptionSubscribe, []byte(""))
	if err != nil {
		return false, err
	}

	// nothing wrong closing socket twice
	var exitTimeout int32

	go func() {
		time.Sleep(timeout)
		atomic.AddInt32(&exitTimeout, 1)
		socket.Close()
	}()
	defer socket.Close()

	for {
		var buf []byte
		if buf, err = socket.Recv(); err != nil {
			if atomic.LoadInt32(&exitTimeout) != 0 {
				return false, nil
			}
			return false, err
		}
		if len(buf) > 0 {
			s := string(buf)

			data := strings.Split(s, " ")
			match := matches(data, pattern)
			if match {
				return true, nil
			}
		}
	}
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

//noinspection GoUnhandledErrorResult
func ListenForPatternMulti(hosts []string, pattern []string, onFinish func(bool), timeout time.Duration, w ...io.Writer) {
	var errout io.Writer
	errout = os.Stdout
	if len(w) != 0 {
		if w[0] == nil {
			errout = ioutil.Discard
		} else {
			errout = w[0]
		}
	}
	var wg sync.WaitGroup
	var counter int32

	wg.Add(len(hosts))

	for _, host := range hosts {
		go func(host string) {
			found, err := subscribeAndWaitForPattern(host, pattern, timeout)
			if err != nil {
				fmt.Fprintf(errout, "[ListenForPatternMulti]: %v\n", err)
			} else {
				if found {
					atomic.AddInt32(&counter, 1)
				}
			}
			wg.Done()
		}(host)
	}
	wg.Wait()

	onFinish(counter == int32(len(hosts)))
}

var ErrorWaitPatternTimeout = errors.New("timeout in SubscribeRunAndWaitForPattern")

func SubscribeRunAndWaitForPattern(hosts []string, topic string, timeout time.Duration, f func() ([]string, error)) error {
	messages := make(chan *HostMessage)

	done := make(chan bool)
	defer func() { done <- true }()

	err := SubscribeMultiOld(hosts, messages, done, topic)
	if err != nil {
		return err
	}

	pattern, err := f()
	if err != nil {
		return err
	}

	received := make(map[string]bool)
	for {
		select {
		case m := <-messages:
			_, ok := received[m.Sender]
			if !ok {
				if matches(m.Message, pattern) {
					received[m.Sender] = true
					if len(received) == len(hosts) {
						return nil
					}
				}
			}

		case <-time.After(timeout):
			return ErrorWaitPatternTimeout
		}
	}
}
