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
var ErrDeviceNotFound = errors.New("NotFound")

type Storable interface {
	GetDevice(string) (string, error)
	UpdateDevice(*UpdateMessage) error
}

type RequestResponse struct {
	IPAddress       string
	DeviceName      string
	Err             error
	ResponseChannel chan RequestResponse
}

func NewGetRequestResponse(ip string) *RequestResponse {
	return &RequestResponse{
		IPAddress:       ip,
		ResponseChannel: make(chan RequestResponse),
	}
}

func NewPostRequestResponse(ip, device string) *RequestResponse {
	return &RequestResponse{
		IPAddress:       ip,
		DeviceName:      device,
		ResponseChannel: make(chan RequestResponse),
	}
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

type DataStore struct {
	SignalChannel chan os.Signal
	GetChannel    chan RequestResponse
	PostChannel   chan RequestResponse
}

func NewDataStore(dataPath string) (ds Storable) {

	datastore := &DataStore{
		SignalChannel: make(chan os.Signal, 1),
		GetChannel:    make(chan RequestResponse),
		PostChannel:   make(chan RequestResponse),
	}
	ds = datastore

	signal.Notify(datastore.SignalChannel, os.Interrupt)
	go datastore.Updater(dataPath)

	return
}

func (ds *DataStore) UpdateDevice(msg *UpdateMessage) error {
	request := NewPostRequestResponse(msg.IPAddress, msg.Device)
	ds.PostChannel <- *request
	response := <-request.ResponseChannel
	return response.Err
}

func (ds *DataStore) GetDevice(ip string) (deviceName string, e error) {

	request := NewGetRequestResponse(ip)
	ds.GetChannel <- *request
	response := <-request.ResponseChannel
	return response.DeviceName, response.Err
}

func (ds *DataStore) Updater(dataLocation string) {

	data := NewData()

	path := filepath.Join(dataLocation, "ipmapper.data")

	if _, err := os.Stat(path); err != nil {
		data.Update(UpdateMessage{IPAddress: "1.2.3.4", Device: "device1"})
		data.Update(UpdateMessage{IPAddress: "1.2.3.5", Device: "device2"})
		data.Update(UpdateMessage{IPAddress: "1.2.3.6", Device: "device3"})
		data.Update(UpdateMessage{IPAddress: "1.2.128.1", Device: "device1"})
		data.Update(UpdateMessage{IPAddress: "1.2.128.2", Device: "device2"})
	} else {
		var buffer []byte
		if buffer, err = ioutil.ReadFile(path); err != nil {
			return
		}

		reader := bytes.NewBuffer(buffer)
		decoder := json.NewDecoder(reader)
		if err = decoder.Decode(data); err != nil {
			log.Println("Could not decode data file", err.Error())
			return
		}

	}

	for {
		select {
		case <-ds.SignalChannel:
			var writer bytes.Buffer
			encoder := json.NewEncoder(&writer)
			if err := encoder.Encode(data); err != nil {
				log.Println(err.Error())
			}

			if err := ioutil.WriteFile(path, writer.Bytes(), 0644); err != nil {
				log.Println(err.Error())
				os.Exit(1)
			}

			os.Exit(0)
		case getRequest := <-ds.GetChannel:
			getRequest.DeviceName, getRequest.Err = data.GetDevice(getRequest.IPAddress)
			getRequest.ResponseChannel <- getRequest
		case postRequest := <-ds.PostChannel:
			u := UpdateMessage{
				IPAddress: postRequest.IPAddress,
				Device:    postRequest.DeviceName,
			}
			postRequest.Err = data.Update(u)
			postRequest.ResponseChannel <- postRequest
		}
	}

}
