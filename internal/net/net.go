package net

import (
	"net"
	"strconv"
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
