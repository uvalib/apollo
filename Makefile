GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOGET=$(GOCMD) get
PWD=$(shell pwd)
BIN=$(PWD)/bin
SRC=$(PWD)/src
VENDOR=$(SRC)/viewsrv/vendor

# project specific definitions
BASE_NAME=serials-manager
SRC_TREE=web
RUNNER=scripts/entry.sh

build: build-darwin build-linux copy-web

all: deps build-darwin build-linux copy-web

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/$(BASE_NAME).darwin $(SRC_TREE)/*.go

copy-web:
	cp -R public/ bin/public/

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/$(BASE_NAME).linux $(SRC_TREE)/*.go

fmt:
	$(GOFMT) $(SRC_TREE)/*

vet:
	$(GOVET) $(SRC_TREE)/*

clean:
	$(GOCLEAN)
	rm -rf $(BIN)

run:
	rm -f $(BIN)/$(BASE_NAME)
	ln -s $(BIN)/$(BASE_NAME).darwin $(BIN)/$(BASE_NAME)
	$(RUNNER)

prep:
	rm -f $(VENDOR)
	rm -f $(SRC)
	ln -s $(PWD)/cmd $(SRC)
	ln -s $(PWD)/vendor $(VENDOR)

deps:
	dep ensure
	dep status
