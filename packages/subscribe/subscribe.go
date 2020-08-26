package subscribe

import (
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

func SubscribeMulti(hosts []string, messages chan<- *HostMessage, done chan bool, topics ...string) error {
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

func waitForPattern(host string, pattern []string, timeout time.Duration) (bool, error) {
	if len(pattern) == 0 {
		return false, fmt.Errorf("wrong pattern")
	}
	socket, err := sub.NewSocket()
	if err != nil {
		return false, err
	}
	for {
		err = socket.Dial("tcp://" + host)
		if err != nil {
			return false, fmt.Errorf("can't dial on sub socket %s: %s", host, err.Error())
		}
		break
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
			found, err := waitForPattern(host, pattern, timeout)
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
