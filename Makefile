GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

build: build-darwin build-linux build-web

all: deps build-darwin build-linux build-web

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/apolloingest.darwin cmd/apolloingest/*.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/apollosvr.darwin cmd/apollosvr/*.go

build-web:
	mkdir -p bin/
	cd frontend/; npm run build

deploy-web:
	mv frontend/dist bin/public

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/apolloingest.linux cmd/apolloingest/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/apollosvr.linux cmd/apollosvr/*.go

clean:
	$(GOCLEAN)
	rm -rf bin

deps:
	dep ensure
	dep status
