package portforward

import (
	"io"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/util/httpstream"
)

// intefaces from the httpstream package used to generate mocks of these dependencies
//
//go:generate mockery --name=Connection --output=automock --outpkg=automock --case=underscore
type Connection interface {
	CreateStream(headers http.Header) (httpstream.Stream, error)
	Close() error
	CloseChan() <-chan bool
	SetIdleTimeout(timeout time.Duration)
	RemoveStreams(streams ...httpstream.Stream)
}

//go:generate mockery --name=Stream --output=automock --outpkg=automock --case=underscore
type Stream interface {
	io.ReadWriteCloser
	Reset() error
	Headers() http.Header
	Identifier() uint32
}
