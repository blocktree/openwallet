.PHONY: all clean
.PHONY: wmd
.PHONY: deps

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

GO ?= latest

# wmd
WMDVERSION = $(shell git describe --tags `git rev-list --tags --max-count=1`)
WMDBINARY = wmd
WMDMAIN = cmd/wmd/main.go

BUILDDIR = build
GITREV = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +'%Y-%m-%d_%T')

WMDLDFLAGS="-X github.com/blocktree/openwallet/cmd/wmd/commands.Version=${WMDVERSION} \
	-X github.com/blocktree/openwallet/cmd/wmd/commands.GitRev=${GITREV} \
	-X github.com/blocktree/openwallet/cmd/wmd/commands.BuildTime=${BUILDTIME}"

# OS platfom
# options: windows-6.0/*,darwin-10.10/amd64,linux/amd64,linux/386,linux/arm64,linux/mips64, linux/mips64le
TARGETS="darwin-10.10/amd64,linux/amd64"

deps:
	go get -u github.com/gythialy/xgo

build:
	GO111MODULE=on go build -ldflags $(WMDLDFLAGS) -i -o $(shell pwd)/$(BUILDDIR)/$(WMDBINARY) $(shell pwd)/$(WMDMAIN)
	@echo "Build $(WMDBINARY) done."


all: wmd

clean:
	rm -rf $(shell pwd)/$(BUILDDIR)/

wmd:
	xgo --dest=$(BUILDDIR) --ldflags=$(WMDLDFLAGS) --out=$(WMDBINARY)-$(WMDVERSION)-$(GITREV) --targets=$(TARGETS) \
	--pkg=$(WMDMAIN) .
