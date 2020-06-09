package publisher

import (
	"fmt"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/sub"
	"sync"
)

func startReadingAsync(url string, topics []string, callback func(msg string)) (func(), error) {
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
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-cancelCh
		socket.Close()
		wg.Done()
	}()

	go func() {
		var msg []byte
		for {
			if msg, err = socket.Recv(); err != nil {
				return
			}
			if len(msg) > 0 {
				callback(string(msg))
			}
		}
	}()
	return func() {
		close(cancelCh)
		wg.Wait()
	}, nil
}

func StartReadingMulti(pubHosts []string, topics []string, chOut chan string) (int, func()) {
	chMulti := make(chan string)
	cancels := make([]func(), 0, len(pubHosts))
	for i, host := range pubHosts {
		cancel, err := startReadingAsync(host, topics, func(msg string) {
			chMulti <- fmt.Sprintf("#%d %s", i, msg)
		})
		if err != nil {
			fmt.Printf("publisher.startReadingAsync %s: %v\n", host, err)
		} else {
			cancels = append(cancels, cancel)
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for msg := range chMulti {
			if len(msg) > 0 {
				chOut <- msg
			}
		}
		wg.Done()
	}()

	return len(cancels), func() {
		for _, cancelFun := range cancels {
			cancelFun()
		}
		close(chMulti)
		wg.Wait()
	}
}
