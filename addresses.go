package main

import (
	"errors"
	"strconv"
	"strings"
)

func HostAndPortFromString(addr, separator string) (string, int, error) {
	addressParts := strings.Split(addr, separator)
	if len(addressParts) != 2 {
		return "", 0, errors.New("invalid address")
	}

	host := addressParts[0]
	port, err := strconv.Atoi(addressParts[1])

	return host, port, err
}
