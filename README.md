# ipmapper


## Install And Run

Requires Go 1.4

Operating systems Linux and OSX

```shell
tar -xzf ipmapper.tgz
IPALLOC_DATAPATH=$(pwd) go run main.go data_store.go validations.go
```

Note that IPALLOC_DATAPATH references the path where the program will store its data file,
ipmapper.data which is used to persist changes when the program terminates.

## Sample Commands

```shell
curl -XPOST http://localhost:8080/addresses/assign -d '{"ip":"1.2.3.34", "device":"deviceXY"}'
curl -XGET http://localhost:8080/devices/1.2.3.34
```

## Tests

Assuming source files are located in $GOPATH/src/ipmapper

```shell
go test ipmapper
```
