package main

import (
	"errors"
	"strconv"
	"strings"
)

type HostAndPort struct {
	Host string
	Port int
}

func HostAndPortFromString(addr, separator string) (HostAndPort, error) {
	addressParts := strings.Split(addr, separator)
	if len(addressParts) != 2 {
		return HostAndPort{}, errors.New("invalid address")
	}

	host := addressParts[0]
	port, err := strconv.Atoi(addressParts[1])

	return HostAndPort{
		Host: host,
		Port: port,
	}, err
}
