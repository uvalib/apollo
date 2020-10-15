GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

build: darwin-srv deploy-templates web

linux-full: linux-srv deploy-templates web

all: darwin-srv linux-srv web deploy-templates

darwin-srv:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/apollosvr.darwin backend/*.go

deploy-templates:
	mkdir -p bin/
	rm -rf bin/templates
	mkdir -p bin/templates
	cp ./templates/* bin/templates

web:
	mkdir -p bin/
	cd frontend/; yarn install; yarn build
	rm -rf bin/public
	mv frontend/dist bin/public

linux-srv:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/apollosvr.linux backend/*.go

clean:
	$(GOCLEAN) ./backend/...
	rm -rf bin

dep:
	cd frontend && yarn upgrade
	$(GOGET) -u ./backend/btsrv/...
	$(GOMOD) tidy
	$(GOMOD) verify