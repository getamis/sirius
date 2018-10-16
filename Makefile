# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

PHONY += all docker test clean

CURDIR := $(shell pwd)
GOBIN := $(shell pwd)/build/bin

TARGETS := $(sort $(notdir $(wildcard ./cmd/*)))
PHONY += $(TARGETS)

all: $(TARGETS)

.SECONDEXPANSION:
$(TARGETS): $(addprefix $(GOBIN)/,$$@)

$(GOBIN):
	@mkdir -p $@

$(GOBIN)/%: $(GOBIN) FORCE
	@go build -v -o $@ ./cmd/$(notdir $@)
	@echo "Done building."
	@echo "Run \"$(subst $(CURDIR),.,$@)\" to launch $(notdir $@)."

coverage.txt:
	@touch $@

test: coverage.txt FORCE
	@for d in `go list ./... | grep -v vendor | grep -v mock`; do		\
		go test -v -coverprofile=profile.out -covermode=atomic $$d;	\
		if [ $$? -eq 0 ]; then						\
			echo "\033[32mPASS\033[0m:\t$$d";			\
			if [ -f profile.out ]; then				\
				cat profile.out >> coverage.txt;		\
				rm profile.out;					\
			fi							\
		else								\
			echo "\033[31mFAIL\033[0m:\t$$d";			\
			exit -1;						\
		fi								\
	done;

# .proto files
PROTOS := \
        health/*.proto

PROTOC_INCLUDES := \
		-I$(CURDIR)/vendor/github.com/golang/protobuf/ptypes \
		-I$(CURDIR)/vendor/github.com/golang/protobuf/ptypes/any \
		-I$(CURDIR)/vendor/github.com/golang/protobuf/ptypes/struct \
		-I$(CURDIR)/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I$(GOPATH)/src

grpc: FORCE
	@protoc $(PROTOC_INCLUDES) \
		--gofast_out=plugins=grpc:$(GOPATH)/src $(addprefix $(CURDIR)/,$(PROTOS))
	@protoc $(PROTOC_INCLUDES) \
		--grpc-gateway_out=logtostderr=true:$(GOPATH)/src $(addprefix $(CURDIR)/,$(PROTOS))

deps:
	docker pull mysql:5.7
	docker pull quay.io/coreos/etcd:v3.0.6
	docker pull amazon/dynamodb-local:latest
	docker pull rabbitmq:3.6.2-management
	docker pull redis:3-alpine

clean:
	rm -fr $(GOBIN)/*

PHONY: help
help:
	@echo  'Generic targets:'
	@echo  '  all                   - Build all targets marked with [*]'
	@echo  '* health                - Build health client'
	@echo  ''
	@echo  'Protobuf targets:'
	@echo  '  grpc                  - Generate gRPC go bindings from .proto files'
	@echo  ''
	@echo  'Test targets:'
	@echo  '  test                  - Run all unit tests'
	@echo  ''
	@echo  'Cleaning targets:'
	@echo  '  clean                 - Remove built executables'
	@echo  ''
	@echo  'Execute "make" or "make all" to build all targets marked with [*] '
	@echo  'For further info see the ./README.md file'

.PHONY: FORCE
FORCE:
