// +build !go1.5
package main

import "testing"

func TestIPInRange(t *testing.T) {
	err := IPInRange(AddressRange, "1.2.1.2")
	if err != nil {
		t.Error("Unexpected", err.Error())
	}

	err = IPInRange(AddressRange, "1.3.1.2")
	if err != ErrIPNotInRange {
		t.Error("Should be", ErrIPNotInRange.Error())
	}

	err = IPInRange(AddressRange, "bob")
	if err == nil {
		t.Error("Garbage IP addy should error")
	}
}

func TestValidateDeviceName(t *testing.T) {
	if !DeviceNameValid("aDevice2") {
		t.Error("Device name should be valid")
	}

	if DeviceNameValid("device Name") {
		t.Error("Device name should not be valid")
	}

	if DeviceNameValid("deviceName ") {
		t.Error("Device name should not be valid")
	}

}
