PKG=github.com/dullgiulio/sima
BINDIR=bin
BINS=pingo
PLUGINS=pingo-hello-world pingo-sleep
PKGDEPS=

all: clean vet fmt build

build: libsima $(BINS) $(EXAMPLES)

fmt:
	go fmt $(PKG)/...

vet:
	go vet $(PKG)/...

libsima:
	go build $(RACE) $(PKG)

clean:
	rm -rf $(BINDIR)/plugins/*
	rm -rf $(BINDIR)/*

bindir:
	mkdir -p $(BINDIR)

bindirex:
	mkdir -p $(BINDIR)/plugins

$(BINS): bindir
	go build $(RACE) -o $(BINDIR)/$@ $(PKG)/examples/$@

$(PLUGINS): bindirex
	go build $(RACE) -o $(BINDIR)/examples/$@ $(PKG)/examples/$@

$(PKGDEPS):
	go get -u $@

.PHONY: all deps build clean fmt vet $(BINS) $(EXAMPLES) $(PKGDEPS)
