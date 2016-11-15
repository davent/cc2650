#export GOOS=linux
#export GOARCH=amd64
BUILD_DIR=./bin

.DEFAULT_GOAL := all

temp:
	go build -o ${BUILD_DIR}/$@ examples/$@.go
	sudo setcap 'cap_net_raw,cap_net_admin=eip' ${BUILD_DIR}/$@

battery:
	go build -o ${BUILD_DIR}/$@ examples/$@.go
	sudo setcap 'cap_net_raw,cap_net_admin=eip' ${BUILD_DIR}/$@

all: temp battery
