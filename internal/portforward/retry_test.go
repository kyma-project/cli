package portforward

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_onErrorRetryTransport_RoundTrip(t *testing.T) {
	testReqBody := "test"

	type fields struct {
		inner http.RoundTripper
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *http.Response
		wantErr error
	}{
		{
			name: "should return response",
			fields: fields{
				inner: &fakeTransport{
					responses: []fakeResponse{
						{
							response: &http.Response{},
							err:      nil,
						},
					},
				},
			},

			args: args{
				req: &http.Request{
					Body: io.NopCloser(bytes.NewReader([]byte(testReqBody))),
				},
			},

			want:    &http.Response{},
			wantErr: nil,
		},
		{
			name: "should return response after 5 retries",
			fields: fields{
				inner: &fakeTransport{
					responses: []fakeResponse{
						{
							response: nil,
							err:      errors.New("test error"),
						},
						{
							response: nil,
							err:      errors.New("test error"),
						},
						{
							response: nil,
							err:      errors.New("test error"),
						},
						{
							response: nil,
							err:      errors.New("test error"),
						},
						{
							response: &http.Response{
								StatusCode: 123,
							},
							err: nil,
						},
					},
				},
			},

			args: args{
				req: &http.Request{
					Body: io.NopCloser(bytes.NewReader([]byte(testReqBody))),
				},
			},

			want: &http.Response{
				StatusCode: 123,
			},
			wantErr: nil,
		},
		{
			name: "should return error",
			fields: fields{
				inner: &fakeTransport{
					responses: []fakeResponse{
						{
							response: nil,
							err:      errors.New("test error"),
						},
						{
							response: nil,
							err:      errors.New("test error"),
						},
						{
							response: nil,
							err:      errors.New("test error"),
						},
						{
							response: nil,
							err:      errors.New("test error"),
						},
						{
							response: nil,
							err:      errors.New("test error"),
						},
					},
				},
			},

			args: args{
				req: &http.Request{
					Body: io.NopCloser(bytes.NewReader([]byte(testReqBody))),
				},
			},

			want: nil,
			wantErr: errors.Join(
				errors.New("test error"),
				errors.New("test error"),
				errors.New("test error"),
				errors.New("test error"),
				errors.New("test error"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &onErrorRetryTransport{
				inner: tt.fields.inner,
			}

			got, err := tr.RoundTrip(tt.args.req)

			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_copyRequest(t *testing.T) {
	testReqBody := "test"

	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *http.Request
		wantErr error
	}{
		{
			name: "should return copied request",
			args: args{
				req: (&http.Request{
					Method: "POST",
					URL: &url.URL{
						Host: "http://test.com",
					},
				}).WithContext(context.Background()),
			},
			want: (&http.Request{
				Method: "POST",
				URL: &url.URL{
					Host: "http://test.com",
				},
			}).WithContext(context.Background()),
			wantErr: nil,
		},
		{
			name: "should return copied request with body",
			args: args{
				req: (&http.Request{
					Method: "POST",
					URL: &url.URL{
						Host: "http://test.com",
					},
					Body: io.NopCloser(bytes.NewReader([]byte(testReqBody))),
				}).WithContext(context.Background()),
			},
			want: (&http.Request{
				Method: "POST",
				URL: &url.URL{
					Host: "http://test.com",
				},
				Body: io.NopCloser(bytes.NewReader([]byte(testReqBody))),
			}).WithContext(context.Background()),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := copyRequest(tt.args.req)

			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.want, got)
		})
	}
}

type fakeResponse struct {
	response *http.Response
	err      error
}

// fakeTransport is a fake http.RoundTripper that returns responses in order
type fakeTransport struct {
	iter      int
	responses []fakeResponse
}

func (t *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res := t.responses[t.iter]
	t.iter++

	return res.response, res.err
}
