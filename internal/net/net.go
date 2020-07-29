package net

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

func GetAvailablePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	_, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return 0, err
	}
	return port, err
}

func DoGet(url string) (int, error) {
	httpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("cannot create a new HTTP request: %v", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cannot send HTTP request to %s: %v", url, err)
	}
	return resp.StatusCode, nil
}
