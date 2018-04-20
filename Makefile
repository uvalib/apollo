GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOGET=$(GOCMD) get

build: build-darwin build-linux copy-web

all: deps build-darwin build-linux copy-web

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/apolloingest.darwin cmd/apolloingest/*.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/apollosvr.darwin cmd/apollosvr/*.go

copy-web:
	cp -R public/ bin/public/

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/apolloingest.linux cmd/apolloingest/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/apollosvr.linux cmd/apollosvr/*.go

fmt:
	$(GOFMT) $(SRC_TREE)/*

vet:
	$(GOVET) $(SRC_TREE)/*

clean:
	$(GOCLEAN)
	rm -rf $(BIN)

deps:
	dep ensure
	dep status
