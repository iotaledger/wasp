// program subscribes and listens to the nanomsg stream publiched by the Wasp host
// and displays it in the console
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/iotaledger/wasp/packages/subscribe"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: submsg <pub host>\n")
		os.Exit(1)
	}
	chMsg := make(chan []string)
	chDone := make(chan bool)
	fmt.Printf("dialing %s\n", os.Args[1])
	err := subscribe.Subscribe(os.Args[1], chMsg, chDone, true, "")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("reading from %s\n", os.Args[1])

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msgSplit := range chMsg {
			fmt.Printf("%s\n", strings.Join(msgSplit, " "))
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) //nolint:govet,staticcheck // TODO check: sigchanyzer: misuse of unbuffered os.Signal channel as argument to signal.Notify

	go func() {
		<-c
		fmt.Printf("interrupt received..\n")
		close(chDone)
	}()

	wg.Wait()
}
