GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOVET = $(GOCMD) vet
GOFMT = $(GOCMD) fmt
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
	cd frontend/; npm install && npm run build
	rm -rf bin/public
	mv frontend/dist bin/public

linux-srv:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/apollosvr.linux backend/*.go

clean:
	$(GOCLEAN) ./backend/...
	rm -rf bin

dep:
	cd frontend && npm upgrade
	$(GOGET) -u ./backend/...
	$(GOMOD) tidy
	$(GOMOD) verify

fmt:
	cd backend; $(GOFMT)

vet:
	cd backend; $(GOVET)

check:
	go install honnef.co/go/tools/cmd/staticcheck
	$(HOME)/go/bin/staticcheck -checks all,-S1002,-ST1003,-S1007,-S1008 backend/*.go
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	$(GOVET) -vettool=$(HOME)/go/bin/shadow ./backend/...
