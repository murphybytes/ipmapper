// +build !go1.5
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	BindAddress  = "127.0.0.1"
	ListenPort   = 8080
	AddressRange = "1.2.0.0/16"
)

var DataFileLocation string
var ErrMethodUnsupported = errors.New("HTTP Method is not supported")
var ErrInvalidPost = errors.New("Expected post body was missing or malformed")

func init() {
	DataFileLocation = os.Getenv("IPALLOC_DATAPATH")
	if DataFileLocation == "" {
		panic("IPALLOC_DATAPATH environment variable must be set to path containing app data")
	}
}

func main() {
	dataStore := NewDataStore(DataFileLocation)

	mux := http.NewServeMux()
	mux.Handle("/devices/", http.StripPrefix("/devices/", NewGetHandler(dataStore)))
	mux.Handle("/addresses/assign", NewPostHandler(dataStore))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", BindAddress, ListenPort), mux)
	if err != nil {
		log.Fatal(err.Error())
	}

}

// ErrorMessage json returned to caller for non 200 requests
type ErrorMessage struct {
	Error string `json:"error"`
	IP    string `json:"ip"`
}

func WriteResponse(w http.ResponseWriter, response interface{}, httpStatus int) {
	var writer bytes.Buffer
	encoder := json.NewEncoder(&writer)
	if err := encoder.Encode(response); err != nil {
		log.Println("Error writing response", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(httpStatus)
	w.Write(writer.Bytes())
}

///////////////////////////////////////////////////////////////////
// Get device
///////////////////////////////////////////////////////////////////

// GetHandler handles get device /devices/<ipaddress>
type GetHandler struct {
	DataStore Storable
}

func NewGetHandler(dataStore Storable) http.Handler {
	return &GetHandler{
		DataStore: dataStore,
	}
}

func (h *GetHandler) Validate(w http.ResponseWriter, r *http.Request) (e error) {
	if r.Method != http.MethodGet {
		e = ErrMethodUnsupported

		response := ErrorMessage{
			Error: fmt.Sprintf("Method=%s %s", r.Method, e.Error()),
			IP:    r.URL.Path,
		}

		WriteResponse(w, response, http.StatusBadRequest)

	}

	return

}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	err := h.Validate(w, r)
	if err != nil {
		return
	}

	ipAddress := r.URL.Path
	deviceName, err := h.DataStore.GetDevice(ipAddress)

	var response interface{}
	var status = http.StatusOK

	if err == nil {
		response = struct {
			Device    string `json:"device"`
			IPAddress string `json:"ip"`
		}{
			IPAddress: ipAddress,
			Device:    deviceName,
		}

	} else {
		response = ErrorMessage{
			Error: err.Error(),
			IP:    ipAddress,
		}

		status = http.StatusNotFound
	}

	WriteResponse(w, response, status)
}

///////////////////////////////////////////////////////////////////
// Assign Address
///////////////////////////////////////////////////////////////////

// UpdateMessage json received in POSTs
type UpdateMessage struct {
	IPAddress string `json:"ip"`
	Device    string `json:"device"`
}

type PostHandler struct {
	DataStore Storable
}

func NewPostHandler(ds Storable) http.Handler {
	return &PostHandler{
		DataStore: ds,
	}
}

func (h *PostHandler) Validate(w http.ResponseWriter, r *http.Request) (msg *UpdateMessage, e error) {

	var updateMessage UpdateMessage

	if r.Method != http.MethodPost {
		e = ErrMethodUnsupported
		response := ErrorMessage{
			IP:    "<unknown>",
			Error: e.Error(),
		}
		WriteResponse(w, response, http.StatusBadGateway)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if e = decoder.Decode(&updateMessage); e != nil {
		response := ErrorMessage{
			IP:    "<unknown>",
			Error: ErrInvalidPost.Error(),
		}
		WriteResponse(w, response, http.StatusBadRequest)
		return
	}

	if e = IPInRange(AddressRange, updateMessage.IPAddress); e != nil {
		response := ErrorMessage{
			IP:    updateMessage.IPAddress,
			Error: e.Error(),
		}
		WriteResponse(w, response, http.StatusBadRequest)
		return
	}

	if !DeviceNameValid(updateMessage.Device) {
		e = fmt.Errorf("Device name is invalid '%s'", updateMessage.Device)
		response := ErrorMessage{
			IP:    updateMessage.IPAddress,
			Error: e.Error(),
		}
		WriteResponse(w, response, http.StatusBadRequest)
		return
	}

	msg = &updateMessage

	return
}

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	if msg, err := h.Validate(w, r); err == nil {
		err = h.DataStore.UpdateDevice(msg)

		var response interface{}
		var status = http.StatusOK

		if err == nil {
			response = struct {
				IPAddress string `json:"ip"`
				Device    string `json:"device"`
			}{
				IPAddress: msg.IPAddress,
				Device:    msg.Device,
			}
			status = http.StatusCreated
		} else {
			response = ErrorMessage{
				Error: err.Error(),
				IP:    msg.IPAddress,
			}

			if err == ErrIPInUse {
				status = http.StatusConflict
			} else {
				status = http.StatusInternalServerError
			}

		}

		WriteResponse(w, response, status)

	}

}
