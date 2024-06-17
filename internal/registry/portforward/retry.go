package portforward

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// onErrorRetryTransport is a RoundTripper that retries requests on error
type onErrorRetryTransport struct {
	inner http.RoundTripper
}

// NewOnErrRetryTransport creates a new onErrorRetryTransport
func NewOnErrRetryTransport(inner http.RoundTripper) http.RoundTripper {
	return &onErrorRetryTransport{
		inner: inner,
	}
}

func (t *onErrorRetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var errList []error
	for i := 0; i < 5; i++ {
		copyReq, err := copyRequest(req)
		if err != nil {
			return nil, fmt.Errorf("error copying request: %v", err)
		}

		resp, err := t.inner.RoundTrip(copyReq)
		if err == nil {
			return resp, nil
		}

		errList = append(errList, err)
	}

	return nil, errors.Join(errList...)
}

// copyRequest copies the request to avoid closing the original request body
func copyRequest(req *http.Request) (*http.Request, error) {
	copy := req.Clone(req.Context())

	if req.Body == nil {
		// stop if body is empty
		return copy, nil
	}

	commonBodyData, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	copy.Body = io.NopCloser(bytes.NewReader(commonBodyData))
	req.Body = io.NopCloser(bytes.NewReader(commonBodyData))

	return copy, nil

}
