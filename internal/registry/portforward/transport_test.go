package portforward

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/kyma-project/cli.v3/internal/registry/portforward/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
)

func Test_portforwardTransport_dialRemote(t *testing.T) {
	t.Run("forward request to stream", func(t *testing.T) {
		testRequest, err := http.NewRequest("GET", "v2", nil)
		require.NoError(t, err)

		testResponse := http.Response{
			Status:     "418 I'm a teapot",
			StatusCode: 418,
		}

		pft := &portforwardTransport{
			remotePort: "8080",
			remoteConn: fixConnectionMock(t,
				fixErrStreamMock(t),
				fixDataStreamMock(t, testRequest, &testResponse),
			),
		}

		got, err := pft.dialRemote(testRequest)

		require.NoError(t, err)
		require.Equal(t, 418, got.StatusCode)
		require.Equal(t, "418 I'm a teapot", got.Status)
	})

	t.Run("errorStream error", func(t *testing.T) {
		testRequest, err := http.NewRequest("GET", "v2", nil)
		require.NoError(t, err)

		testResponse := http.Response{
			Status:     "418 I'm a teapot",
			StatusCode: 418,
		}

		pft := &portforwardTransport{
			remotePort: "8080",
			remoteConn: fixConnectionMock(t,
				fixBrokenErrStreamMock(t, "", errors.New("test error")),
				fixDataStreamMock(t, testRequest, &testResponse),
			),
		}

		got, err := pft.dialRemote(testRequest)

		require.EqualError(t, err, "error reading from error stream: test error")
		require.Equal(t, 418, got.StatusCode)
		require.Equal(t, "418 I'm a teapot", got.Status)
	})

	t.Run("errorStream error message", func(t *testing.T) {
		testRequest, err := http.NewRequest("GET", "v2", nil)
		require.NoError(t, err)

		testResponse := http.Response{
			Status:     "418 I'm a teapot",
			StatusCode: 418,
		}

		pft := &portforwardTransport{
			remotePort: "8080",
			remoteConn: fixConnectionMock(t,
				fixBrokenErrStreamMock(t, "test error message", io.EOF),
				fixDataStreamMock(t, testRequest, &testResponse),
			),
		}

		got, err := pft.dialRemote(testRequest)

		require.EqualError(t, err, "an error occurred while forwarding: test error message")
		require.Equal(t, 418, got.StatusCode)
		require.Equal(t, "418 I'm a teapot", got.Status)
	})

	t.Run("dataStream write error", func(t *testing.T) {
		testRequest, err := http.NewRequest("GET", "v2", nil)
		require.NoError(t, err)

		pft := &portforwardTransport{
			remotePort: "8080",
			remoteConn: fixConnectionMock(t,
				fixErrStreamMock(t),
				fixBrokenDataStreamMock(t, errors.New("test error"), nil),
			),
		}

		got, err := pft.dialRemote(testRequest)

		require.EqualError(t, err, "error writing to remote stream: test error")
		require.Nil(t, got)
	})

	t.Run("dataStream read error", func(t *testing.T) {
		testRequest, err := http.NewRequest("GET", "v2", nil)
		require.NoError(t, err)

		pft := &portforwardTransport{
			remotePort: "8080",
			remoteConn: fixConnectionMock(t,
				fixErrStreamMock(t),
				fixBrokenDataStreamMock(t, nil, errors.New("test error")),
			),
		}

		got, err := pft.dialRemote(testRequest)

		require.EqualError(t, err, "error reading from remote stream: malformed HTTP response \"\\x00\"")
		require.Nil(t, got)
	})
}

func fixConnectionMock(t *testing.T, errStream, dataStream httpstream.Stream) httpstream.Connection {
	connMock := automock.NewConnection(t)
	connMock.On("RemoveStreams", errStream).Once()
	connMock.On("CreateStream", mock.MatchedBy(func(header http.Header) bool {
		if header.Get(v1.StreamType) == v1.StreamTypeError &&
			header.Get(v1.PortHeader) == "8080" &&
			header.Get(v1.PortForwardRequestIDHeader) != "0" {
			return true
		}

		return false
	})).Return(errStream, nil).Once()

	connMock.On("RemoveStreams", dataStream).Once()
	connMock.On("CreateStream", mock.MatchedBy(func(header http.Header) bool {
		if header.Get(v1.StreamType) == v1.StreamTypeData &&
			header.Get(v1.PortHeader) == "8080" &&
			header.Get(v1.PortForwardRequestIDHeader) != "0" {
			return true
		}

		return false
	})).Return(dataStream, nil).Once()

	return connMock
}

func fixErrStreamMock(t *testing.T) httpstream.Stream {
	errStreamMock := automock.NewStream(t)
	errStreamMock.On("Close").Return(nil).Once()
	errStreamMock.On("Read", mock.Anything).Return(0, io.EOF).Once()
	return errStreamMock
}

func fixBrokenErrStreamMock(t *testing.T, message string, err error) httpstream.Stream {
	errStreamMock := automock.NewStream(t)
	errStreamMock.On("Close").Return(nil).Once()
	errStreamMock.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		require.True(t, ok)

		copy(b, []byte(message))
	}).Return(len(message), err).Once()
	return errStreamMock
}

func fixDataStreamMock(t *testing.T, req *http.Request, resp *http.Response) httpstream.Stream {
	reqBuf := bytes.NewBuffer([]byte{})
	err := req.Write(reqBuf)
	require.NoError(t, err)

	reqBytes := reqBuf.Bytes()

	responseBuf := bytes.NewBuffer([]byte{})
	resp.Write(responseBuf)

	dataStreamMock := automock.NewStream(t)
	dataStreamMock.On("Close").Return(nil).Twice()
	dataStreamMock.On("Write", mock.MatchedBy(func(req []byte) bool {
		return bytes.Equal(req, reqBytes)
	})).Return(len(reqBytes), nil).Once()
	dataStreamMock.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		b, ok := args.Get(0).([]byte)
		require.True(t, ok)

		copy(b, responseBuf.Bytes())
	}).Return(responseBuf.Len(), nil).Once()
	return dataStreamMock
}

func fixBrokenDataStreamMock(t *testing.T, writeErr, readErr error) httpstream.Stream {
	dataStreamMock := automock.NewStream(t)
	dataStreamMock.On("Close").Return(nil)
	dataStreamMock.On("Write", mock.Anything).Return(4096, writeErr) // 4096 is the max buffer size
	dataStreamMock.On("Read", mock.Anything).Return(1, readErr).Maybe()
	return dataStreamMock
}
