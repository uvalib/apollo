GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

build: darwin-srv deploy-templates web

linux-service: linux-srv deploy-templates web

all: darwin-srv linux-srv darwin-tools linux-tools web deploy-templates

darwin-tools:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/apolloingest.darwin backend/cmd/apolloingest/*.go

darwin-srv:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/apollosvr.darwin backend/cmd/apollosvr/*.go

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

linux-tools:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/apolloingest.linux backend/cmd/apolloingest/*.go

linux-srv:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/apollosvr.linux backend/cmd/apollosvr/*.go

clean:
	rm -rf bin
