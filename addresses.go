package main

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

type HostAndPort struct {
	Host string
	Port int
}

func (hp *HostAndPort) String() string {
	return hp.Host + ":" + strconv.Itoa(hp.Port)
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

func GetCanonicalHostAndPort(addr, separator string) (HostAndPort, error) {
	hostAndPort, err := HostAndPortFromString(addr, separator)
	if err != nil {
		return hostAndPort, err
	}

	hostAndPort.Host, err = GetCanonicalHostname(hostAndPort.Host)
	if err != nil {
		return hostAndPort, err
	}

	return hostAndPort, nil
}

func GetCanonicalHostname(hostname string) (string, error) {
	names, err := net.LookupAddr(hostname)
	if err != nil {
		return "", err
	}

	return names[0][:len(names[0])-1], nil
}
