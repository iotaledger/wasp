package publisher

import (
	"fmt"
	"strings"

	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/sub"
	_ "go.nanomsg.org/mangos/v3/transport/all"
)

func Subscribe(host string, messages chan<- []string, done <-chan bool, topics ...string) error {
	socket, err := sub.NewSocket()
	if err != nil {
		return err
	}
	if err = socket.Dial("tcp://" + host); err != nil {
		return fmt.Errorf("can't dial on sub socket %s: %s", host, err.Error())
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
		err := Subscribe(host, hostMessages, done, topics...)
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
