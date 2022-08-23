DOCKER ?= podman

PROTOS = tester service-tester

PROTO_DIR = proto

PROTO_FILES = $(foreach proto,$(PROTOS),$(proto).proto)

CUR_DIR = $(shell /bin/pwd)

all: descriptors

descriptors:
	protoc \
		--descriptor_set_out=desc.pb \
		-I$(PROTO_DIR) \
		--include_source_info \
		$(PROTO_FILES)

check: plugin
	cd $(PROTO_DIR) && protoc \
		--docjson_out=. \
		--plugin=$(CUR_DIR)/cmd/protoc-gen-docjson/protoc-gen-docjson \
		-I. \
		$(PROTO_FILES)

plugin: cmd/protoc-gen-docjson/protoc-gen-docjson

cmd/protoc-gen-docjson/protoc-gen-docjson: cmd/protoc-gen-docjson/protoc-gen-docjson.go
	cd cmd/protoc-gen-docjson && go build -a
