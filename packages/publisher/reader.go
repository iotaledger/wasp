package publisher

import (
	"fmt"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/sub"
	"strings"
)

func StartReading(url string, topics []string, callback func(msgType string, parts ...string)) (func(), error) {
	socket, err := sub.NewSocket()
	if err != nil {
		return nil, err
	}
	if err = socket.Dial(url); err != nil {
		return nil, fmt.Errorf("can't dial on sub socket: %s", err.Error())
	}
	err = socket.SetOption(mangos.OptionSubscribe, topics)
	if err != nil {
		return nil, err
	}
	cancelCh := make(chan struct{})
	go func() {
		<-cancelCh
		socket.Close()
	}()

	go func() {
		var msg []byte
		for {
			if msg, err = socket.Recv(); err != nil {
				return
			}
			msgSplit := strings.Split(string(msg), " ")
			if len(msgSplit) != 0 {
				callback(msgSplit[0], msgSplit[1:]...)
			}
		}
	}()
	return func() {
		close(cancelCh)
	}, nil
}
