GO ?= go
CONFIG_DIR ?= undef

# Platforms to build for. You can add or remove platforms here
# Note that at this time, Linux build does not work outside of linux...
LOCAL_PLATFORM := $(shell uname -s | tr A-Z a-z)
PLATFORMS = linux darwin

SOURCE_FILES := $(shell find . -name '*.go')

PWD ?= $(shell pwd)
REAL_CONFIG_DIR := $(shell if [ -d $(CONFIG_DIR) -a ! -d /$(CONFIG_DIR) ]; then echo $(PWD)/$(CONFIG_DIR); else echo $(CONFIG_DIR); fi)
BINARY := $(shell basename $(PWD))
BUILD_DIR := $(PWD)
OUTPUT_BASE_DIR = $(BUILD_DIR)/output/
BUILD_DIR_LINK = $(shell readlink $(BUILD_DIR))

VET_REPORT := vet.report
TEST_REPORT := tests.xml

GOPRIVATE=github.com/AntonioMA

GOARCH := amd64

VERSION?=$(shell grep version Dockerfile| cut -f 2 -d = | tr -d '\n\r')
COMMIT := $(shell git rev-parse HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)


# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=$(VERSION) -X main.COMMIT=$(COMMIT) -X main.BRANCH=$(BRANCH)"

OUTPUT_PLATFORM_DIRS = $(addprefix $(OUTPUT_BASE_DIR), $(PLATFORMS))

# To-do: Check if we can use this
#SUFFIXES+= .go .so
#.SUFFIXES: $(SUFFIXES)

# Those are calculated based on your configuration above
all: showconfig $(OUTPUT_PLATFORM_DIRS) $(PLATFORMS)

-include common-deploy/Makefile

local: showconfig $(OUTPUT_PLATFORM_DIRS) $(LOCAL_PLATFORM)

docker-img: Dockerfile $(SOURCE_FILES) linux
	@docker build . -t antonioma/http-echo:latest -t antonioma/http-echo:$(COMMIT) -t antonioma/http-echo:$(VERSION)

docker-run: docker-img
	@if [ ! -d $(REAL_CONFIG_DIR) ];  then echo "Need a valid config directory to create the image"; exit 1; fi
	docker run -p 8123:8125 --rm --env "PORT=8125" -v $(REAL_CONFIG_DIR):/config http-echo

showconfig:
	@echo Variables:
	@echo PLATFORMS: $(PLATFORMS)
	@echo OUTPUT_PLATFORM_DIRS: $(OUTPUT_PLATFORM_DIRS)
	@echo DEPLOY_TARGETS: $(DEPLOY_TARGETS)

$(OUTPUT_PLATFORM_DIRS):
	@echo "Creating output directories"
	@mkdir -p $@

$(PLATFORMS): OUTPUT_DIR = $(OUTPUT_BASE_DIR)$@/

$(PLATFORMS):
	@echo Building $@ from $(BUILD_DIR)
	@rm -f $(OUTPUT_DIR)$(BINARY)
	@cd $(BUILD_DIR); \
	GO111MODULE=on GOOS=$@ GOARCH=${GOARCH} GOPRIVATE=$(GOPRIVATE) $(GO) build -v ${LDFLAGS} -o $(OUTPUT_DIR)$(BINARY) .

run: clean all
	cd $(OUTPUT_BASE_DIR)$(LOCAL_PLATFORM); \
	./$(BINARY) -configFile ../../samples/config.json

run-delta: clean all
	cd $(OUTPUT_BASE_DIR)$(LOCAL_PLATFORM); \
	./$(BINARY) -configFile ../../samples/config.json | delta -config meigasConfig/deltaConfig.json

test:
	if ! hash go2xunit 2>/dev/null; then $(GO) install github.com/tebeka/go2xunit; fi
	cd ${BUILD_DIR}; \
	$(GODEP) $(GO) test -v ./... 2>&1 | go2xunit -output ${TEST_REPORT}

vet:
	-cd ${BUILD_DIR}; \
	$(GODEP) $(GO) vet ./... > ${VET_REPORT} 2>&1

fmt:
	cd ${BUILD_DIR}; \
	$(GO) fmt $$($(GO) list ./... | grep -v /vendor/)

clean:
	-rm -f ${TEST_REPORT}
	-rm -f ${VET_REPORT}
	-rm -rf $(OUTPUT_BASE_DIR)

.PHONY: all link $(PLATFORMS) test vet fmt clean showconfig $(DEPLOY_TARGETS)

.SECONDEXPANSION:

