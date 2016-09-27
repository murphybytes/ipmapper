// +build !go1.5
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

var ErrIPInUse = errors.New("IP Address is in use by another device")
var ErrDeviceNotFound = errors.New("No device was found for IP address")

type Storable interface {
	Close()
}

type UpdateMessage struct {
	IPAddress string `json:"ip"`
	Device    string `json:"device"`
}

type Data struct {
	IPToDevice  map[string]string   `json:"ip_to_device"`
	DeviceToIPs map[string][]string `json:"device_to_ips"`
}

func NewData() *Data {
	return &Data{
		IPToDevice:  map[string]string{},
		DeviceToIPs: map[string][]string{},
	}
}

type DataStore struct {
	Data          Data
	SignalChannel chan os.Signal
}

func NewDataStore(dataPath string) (ds Storable, e error) {

	dataStore := &DataStore{
		SignalChannel: make(chan os.Signal, 1),
		Data: Data{
			IPToDevice:  map[string]string{},
			DeviceToIPs: map[string][]string{},
		},
	}

	signal.Notify(dataStore.SignalChannel, os.Interrupt)
	go dataStore.Updater()

	path := filepath.Join(DataFileLocation, "ipmapper")

	if _, err := os.Stat(path); err != nil {
		return
	}

	// if _, e = ioutil.ReadFile(path); e != nil {
	// 	log.Println("ReadFile fails ", e.Error())
	// }
	//
	ds = dataStore
	return
}

func (ds *DataStore) Close() {
	log.Println("xxxxxx")
	path := filepath.Join(DataFileLocation, "ipmapper")

	var writer bytes.Buffer
	encoder := json.NewEncoder(&writer)
	if err := encoder.Encode(ds.Data); err != nil {
		log.Println(err.Error())
		return
	}

	if err := ioutil.WriteFile(path, writer.Bytes(), 0644); err != nil {
		log.Println(err.Error())
	}

	log.Println("wrote file to ", path)

}

func (d *Data) GetDevice(ip string) (device string, e error) {
	var ok bool
	if device, ok = d.IPToDevice[ip]; ok {
		return
	}

	e = ErrDeviceNotFound
	return
}

func (d *Data) Update(u UpdateMessage) error {
	// see if ip is mapped to another device
	if device, ok := d.IPToDevice[u.IPAddress]; ok {
		if device != u.Device {
			return ErrIPInUse
		}
	}

	d.IPToDevice[u.IPAddress] = u.Device

	if _, ok := d.DeviceToIPs[u.Device]; !ok {
		d.DeviceToIPs[u.Device] = []string{}
	}

	d.DeviceToIPs[u.Device] = append(d.DeviceToIPs[u.Device], u.IPAddress)

	return nil
}

func (ds *DataStore) Updater() {

	for {
		select {
		case <-ds.SignalChannel:
			ds.Close()
			os.Exit(0)

		}
	}

}
