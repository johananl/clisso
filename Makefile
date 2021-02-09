GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BUILDPATH=build
BINARY_NAME=clisso
VERSION=`git describe --tags --always --dirty`

# Use the Go module mirror - https://blog.golang.org/module-mirror-launch.
export GOPROXY=https://proxy.golang.org

.PHONY: build
build:
	$(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME) -v

.PHONY: test
test:
	$(GOCMD) test -v ./...

.PHONY: darwin-amd64
darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-darwin-amd64 -v

.PHONY: linux-386
linux-386:
	GOOS=linux GOARCH=386 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-linux-386 -v

.PHONY: linux-amd64
linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-linux-amd64 -v

.PHONY: windows-386
windows-386:
	GOOS=windows GOARCH=386 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-windows-386.exe -v

.PHONY: windows-amd64
windows-amd64:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "-X main.version=$(VERSION)" -o $(BUILDPATH)/$(BINARY_NAME)-windows-amd64.exe -v

.PHONY: all
all: darwin-amd64 linux-386 linux-amd64 windows-386 windows-amd64

.PHONY: zip
zip:
	for i in `ls -1 $(BUILDPATH)/$(BINARY_NAME)* | grep -v '.zip'`; do zip $$i.zip $$i; done

.PHONY: brew
brew:
	bash make_brew_release.sh $(BINARY_NAME) $(VERSION) $(BUILDPATH)

.PHONY: release
release: clean all zip brew

.PHONY: install
install:
	go install -ldflags "-X main.version=$(VERSION)"

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BUILDPATH)/$(BINARY_NAME)*
