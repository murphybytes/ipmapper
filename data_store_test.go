// +build !go1.5 testfile
package main

import "testing"

func TestUpdate(t *testing.T) {
	d := NewData()

	um := UpdateMessage{
		IPAddress: "10.2.3.4",
		Device:    "device1",
	}
	if err := d.Update(um); err != nil {
		t.Error("Expected nil error")
	}

	if len(d.DeviceToIPs[um.Device]) != 1 {
		t.Error("Expected one IP address")
	}

	um = UpdateMessage{
		IPAddress: "10.2.3.4",
		Device:    "device2",
	}

	err := d.Update(um)
	if err != ErrIPInUse {
		t.Error("Expected ", ErrIPInUse.Error())
	}

}

func TestGetDevice(t *testing.T) {
	d := NewData()

	um := UpdateMessage{
		IPAddress: "10.2.3.4",
		Device:    "device1",
	}

	d.Update(um)

	um = UpdateMessage{
		IPAddress: "10.2.3.5",
		Device:    "device1",
	}

	d.Update(um)

	device, err := d.GetDevice("10.2.3.4")

	if err != nil {
		t.Error("Unexpected error")
	}

	if device != um.Device {
		t.Error("Expected", um.Device, "Actual", device)
	}

	device, err = d.GetDevice("10.2.3.5")

	if err != nil {
		t.Error("Unexpected error")
	}

	if device != um.Device {
		t.Error("Expected", um.Device, "Actual", device)
	}

}
