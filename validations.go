// +build !go1.5
package main

import (
	"errors"
	"net"
	"regexp"
)

var ErrIPNotInRange = errors.New("IP Address is not in range")
var ErrCIDRNotValid = errors.New("CIDR string is malformed")

var RegEx = regexp.MustCompile("^[a-z0-9A-Z]+$")

func IPInRange(cidr string, ip string) error {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return ErrCIDRNotValid
	}

	if !network.Contains(net.ParseIP(ip)) {
		return ErrIPNotInRange
	}

	return nil

}

func DeviceNameValid(deviceName string) bool {
	s := RegEx.FindString(deviceName)
	return s == deviceName
}
