SHELL=PATH='$(PATH)' /bin/sh


GOBUILD=CGO_ENABLED=0 go build -ldflags '-w -s'

PLATFORM := $(shell uname -o)

NAME := youPipe.exe
ifeq ($(PLATFORM), Msys)
    INCLUDE := ${shell echo "$(GOPATH)"|sed -e 's/\\/\//g'}
else ifeq ($(PLATFORM), Cygwin)
    INCLUDE := ${shell echo "$(GOPATH)"|sed -e 's/\\/\//g'}
else
	INCLUDE := $(GOPATH)
	NAME=youPipe
endif

# enable second expansion
.SECONDEXPANSION:

.PHONY: all
.PHONY: pbs

BINDIR=$(INCLUDE)/bin

all: pbs  build

build:
	GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)

pbs:
	cd pbs/ && $(MAKE)

clean:
	rm $(BINDIR)/$(NAME)
