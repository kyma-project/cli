package cli

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFinalizer_Add(t *testing.T) {
	tests := []struct {
		name        string
		funcs       []func()
		f           func()
		expectedLen int
	}{
		{
			name:        "should add func to the empty array",
			funcs:       []func(){},
			f:           func() {},
			expectedLen: 1,
		},
		{
			name:        "should add func to the nil array",
			funcs:       nil,
			f:           func() {},
			expectedLen: 1,
		},
		{
			name:        "should add nil func to the empty array",
			funcs:       []func(){},
			f:           nil,
			expectedLen: 1,
		},
		{
			name: "should add func to the array",
			funcs: []func(){
				func() {},
				func() {},
				func() {},
			},
			f:           func() {},
			expectedLen: 4,
		},
	}
	for _, tt := range tests {
		expectedLen := tt.expectedLen
		funcs := tt.funcs
		f := tt.f

		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				d := &Finalizers{
					funcs: funcs,
				}

				d.Add(f)

				require.Equal(t, expectedLen, len(d.funcs))
			},
		)
	}
}

func TestFinalizer_setupCloseHandler(t *testing.T) {
	type fields struct {
		notify func(c chan<- os.Signal, sig ...os.Signal)
		funcs  []func(chan int) func()
	}
	tests := []struct {
		name           string
		fields         fields
		funcExecutions int
		nilFuncs       int
	}{
		{
			name: "should receive SIGTERM syscall and run function",
			fields: fields{
				notify: fixNotify(syscall.SIGTERM),
				funcs: []func(chan int) func(){
					fixFunc,
				},
			},
		},
		{
			name: "should receive SIGINT syscall and run all functions",
			fields: fields{
				notify: fixNotify(syscall.SIGINT),
				funcs: []func(chan int) func(){
					fixFunc, fixFunc,
				},
			},
		},
		{
			name: "should end process after timeout will occurred",
			fields: fields{
				notify: fixNotify(syscall.SIGINT),
				funcs: []func(chan int) func(){
					fixFuncWithSleep, fixFuncWithSleep,
					fixFuncWithSleep,
				},
			},
		},
		{
			name: "should receive SIGINT syscall and run all (non nil) functions",
			fields: fields{
				notify: fixNotify(syscall.SIGINT),
				funcs: []func(chan int) func(){
					fixNilFunc, fixFunc, fixNilFunc, fixFunc, fixNilFunc,
				},
			},
			nilFuncs: 3,
		},
	}
	for _, tt := range tests {
		funcs := tt.fields.funcs
		nilFuncs := tt.nilFuncs
		notify := tt.fields.notify

		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				counterChan := make(chan int, len(funcs)-nilFuncs)
				exit := make(chan struct{})

				d := &Finalizers{
					notify: notify,
					exit:   fixExit(exit),
					funcs:  fixFuncs(counterChan, funcs),
				}

				d.setupCloseHandler()

				// wait until all functions end
				for i := len(funcs) - nilFuncs; i != 0; i-- {
					<-counterChan
				}

				<-exit
			},
		)
	}
}

func fixNotify(signal os.Signal) func(c chan<- os.Signal, sig ...os.Signal) {
	return func(c chan<- os.Signal, sig ...os.Signal) {
		time.Sleep(time.Second)
		c <- signal
	}
}

func fixFuncs(counter chan int, functions []func(counter chan int) func()) []func() {
	var fixedFuncs []func()
	for _, f := range functions {
		fixedFuncs = append(fixedFuncs, f(counter))
	}
	return fixedFuncs
}

func fixFunc(counter chan int) func() {
	return func() {
		counter <- 1
	}
}

func fixFuncWithSleep(counter chan int) func() {
	return func() {
		time.Sleep(time.Second * 2)
		counter <- 1
	}
}

func fixNilFunc(_ chan int) func() {
	return nil
}

func fixExit(exit chan struct{}) func(int) {
	return func(_ int) {
		exit <- struct{}{}
	}
}
