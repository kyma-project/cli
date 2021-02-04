package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const timeout = 5

type Finalizer struct {
	notify func(c chan<- os.Signal, sig ...os.Signal)
	exit   func(int)
	funcs  []func()
}

func NewFinalizer() *Finalizer {
	return &Finalizer{
		notify: signal.Notify,
		exit:   os.Exit,
	}
}

func (f *Finalizer) Add(function func()) {
	f.funcs = append(f.funcs, function)
}

func (f *Finalizer) SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	f.notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-c
		lastChan := make(chan struct{})
		go func() {
			for _, f := range f.funcs {
				if f != nil {
					f()
				}
			}
			lastChan <- struct{}{}
		}()

		go func() {
			time.Sleep(time.Second * timeout)
			lastChan <- struct{}{}
		}()

		<-lastChan
		fmt.Printf("\r- Signal '%v' received from Terminal. Exiting...\n ", sig)
		f.exit(0)
	}()
}
