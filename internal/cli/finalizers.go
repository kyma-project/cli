package cli

import (
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const timeout = 5 * time.Second

type Finalizers struct {
	notify func(c chan<- os.Signal, sig ...os.Signal)
	exit   func(int)
	funcs  []func()
	logger zap.SugaredLogger
}

func NewFinalizer() *Finalizers {
	fin := &Finalizers{
		notify: signal.Notify,
		exit:   os.Exit,
		logger: *NewLogger(false).Sugar(),
	}
	fin.setupCloseHandler()
	return fin
}

func (f *Finalizers) Add(function func()) {
	f.funcs = append(f.funcs, function)
}

func (f *Finalizers) setupCloseHandler() {
	wg := sync.WaitGroup{}
	c := make(chan os.Signal, 1)
	f.notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-c
		f.logger.Infof("\r- Signal '%v' received from Terminal. Exiting...\n ", sig)
		wg.Add(1)
		for _, f := range f.funcs {
			if f != nil {
				go func() {
					defer wg.Done()
					f()
				}()
			}
		}
		waitTimeout(&wg, timeout)
		f.exit(0)
	}()
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}