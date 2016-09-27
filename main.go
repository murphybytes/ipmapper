// +build !go1.5
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	BindAddress   = "127.0.0.1"
	ListenPort    = 8888
	AddressRange  = "1.2.0.0/16"
	HighIPAddress = "1.2.255.255"
)

var DataFileLocation string

func init() {
	DataFileLocation = os.Getenv("IPALLOC_DATAPATH")
	if DataFileLocation == "" {
		panic("IPALLOC_DATAPATH environment variable must be set to path containing app data")
	}
}

func main() {
	dataStore, err := NewDataStore(DataFileLocation)
	if err != nil {
		log.Fatal("Could not load data ", err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("/devices/", http.StripPrefix("/devices/", NewGetHandler(dataStore)))
	mux.HandleFunc("/addresses/assign", HandlePostAddressesAssign(dataStore))
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", BindAddress, ListenPort), mux)
	if err != nil {
		log.Fatal(err.Error())
	}

}

type GetHandler struct {
	dataStore Storable
}

func NewGetHandler(dataStore Storable) http.Handler {
	return &GetHandler{
		dataStore: dataStore,
	}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("path", r.URL.Path)
	w.WriteHeader(http.StatusOK)
}

func HandlePostAddressesAssign(dataStore Storable) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
