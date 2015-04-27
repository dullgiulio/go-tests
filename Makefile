PKG=github.com/dullgiulio/sima
BINDIR=bin
BINS=simactl
EXAMPLES=sima-hello-world sima-sleep
PKGDEPS=

all: clean vet fmt build

build: libsima $(BINS) $(EXAMPLES)

fmt:
	go fmt $(PKG)/...

vet:
	go vet $(PKG)/...

libsima:
	go build  $(PKG)

clean:
	rm -rf $(BINDIR)/examples/*
	rm -rf $(BINDIR)/*

bindir:
	mkdir -p $(BINDIR)

bindirex:
	mkdir -p $(BINDIR)/examples

$(BINS): bindir
	go build  -o $(BINDIR)/$@ $(PKG)/cli/$@

$(EXAMPLES): bindirex
	go build  -o $(BINDIR)/examples/$@ $(PKG)/examples/$@

$(PKGDEPS):
	go get -u $@

.PHONY: all deps build clean fmt vet $(BINS) $(EXAMPLES) $(PKGDEPS)
