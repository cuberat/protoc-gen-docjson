DOCKER ?= podman

PROTOS = tester service-tester

PROTO_DIR = proto

PROTO_FILES = $(foreach proto,$(PROTOS),$(PROTO_DIR)/$(proto).proto)

all: descriptors

descriptors:
	protoc \
		--descriptor_set_out=desc.pb \
		-I$(PROTO_DIR) \
		--include_source_info \
		$(PROTO_FILES)
