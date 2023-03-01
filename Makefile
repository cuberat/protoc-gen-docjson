DOCKER ?= podman

TOP_DIR = $(shell /bin/pwd)
OUTFILE = protos.json
READABLE_OUTFILE = protos_readable.json

PROTOS = tester service-tester subdir/docstuff
PROTO_DIR = $(TOP_DIR)/proto
PROTO_FILES = $(foreach proto,$(PROTOS),$(PROTO_DIR)/$(proto).proto)


OUT_DIR = $(TOP_DIR)


all: descriptors

descriptors:
	protoc \
		--descriptor_set_out=desc.pb \
		-I$(PROTO_DIR) \
		--include_source_info \
		$(PROTO_FILES)

check: plugin
	protoc \
		--docjson_out="$(OUT_DIR)" \
		--docjson_opt=outfile=$(OUTFILE),proto=$(PROTO_DIR) \
		--plugin=$(TOP_DIR)/cmd/protoc-gen-docjson/protoc-gen-docjson \
		-I$(PROTO_DIR) \
		$(PROTO_FILES)
	cat $(TOP_DIR)/$(OUTFILE) | jq > $(TOP_DIR)/$(READABLE_OUTFILE)

checkdebug: plugin
	protoc \
		--docjson_out="$(OUT_DIR)" \
		--docjson_opt=outfile=$(OUTFILE),proto=$(PROTO_DIR),debug \
		--plugin=$(TOP_DIR)/cmd/protoc-gen-docjson/protoc-gen-docjson \
		-I$(PROTO_DIR) \
		$(PROTO_FILES)
	cat $(TOP_DIR)/$(OUTFILE) | jq > $(TOP_DIR)/$(READABLE_OUTFILE)

# plugin: cmd/protoc-gen-docjson/protoc-gen-docjson

plugin:
	cd cmd/protoc-gen-docjson && go build -a

# cmd/protoc-gen-docjson/protoc-gen-docjson:
# 	cd cmd/protoc-gen-docjson && go build -a
