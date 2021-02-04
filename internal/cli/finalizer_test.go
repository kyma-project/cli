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

		t.Run(tt.name, func(t *testing.T) {
			d := &Finalizer{
				funcs: funcs,
			}

			d.Add(f)

			require.Equal(t, expectedLen, len(d.funcs))
		})
	}
}

func TestFinalizer_SetupCloseHandler(t *testing.T) {
	type fields struct {
		notify func(c chan<- os.Signal, sig ...os.Signal)
		funcs  []func(*uint64) func()
	}
	tests := []struct {
		name           string
		fields         fields
		funcExecutions uint64
	}{
		{
			name: "should receive SIGINT syscall and run function",
			fields: fields{
				notify: fixNotify(syscall.SIGINT, time.Second),
				funcs: []func(*uint64) func(){
					fixFunc,
				},
			},
			funcExecutions: 1,
		},
		{
			name: "should receive SIGINT syscall and run all functions",
			fields: fields{
				notify: fixNotify(syscall.SIGINT, time.Second),
				funcs: []func(*uint64) func(){
					fixFunc, fixFunc,
				},
			},
			funcExecutions: 2,
		},
		{
			name: "should end process after timeout will occurred",
			fields: fields{
				notify: fixNotify(syscall.SIGINT, time.Second),
				funcs: []func(*uint64) func(){
					fixFunc, fixFunc, fixFunc, fixFunc,
					fixFunc, fixFunc, fixFunc, fixFunc,
				},
			},
			funcExecutions: 2,
		},
	}
	for _, tt := range tests {
		funcs := tt.fields.funcs
		notify := tt.fields.notify
		funcExecution := tt.funcExecutions

		t.Run(tt.name, func(t *testing.T) {
			counter := uint64(0)
			exit := make(chan struct{})
			d := &Finalizer{
				notify: notify,
				exit:   fixExit(exit),
				funcs:  fixFuncs(&counter, funcs),
			}

			d.SetupCloseHandler()

			<-exit
			require.Equal(t, funcExecution, counter)
		})
	}
}

func fixNotify(signal os.Signal, duration time.Duration) func(c chan<- os.Signal, sig ...os.Signal) {
	return func(c chan<- os.Signal, sig ...os.Signal) {
		time.Sleep(duration)
		c <- signal
	}
}

func fixFuncs(counter *uint64, functions []func(counter *uint64) func()) []func() {
	var fixedFuncs []func()
	for _, f := range functions {
		fixedFuncs = append(fixedFuncs, f(counter))
	}
	return fixedFuncs
}

func fixFunc(counter *uint64) func() {
	return func() {
		time.Sleep(time.Second * 2)
		*counter++
	}
}

func fixExit(exit chan struct{}) func(int) {
	return func(_ int) {
		exit <- struct{}{}
	}
}
