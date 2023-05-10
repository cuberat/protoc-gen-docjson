DOCKER ?= podman

TOP_DIR = $(shell /bin/pwd)
OUTFILE = protos.json
READABLE_OUTFILE = protos_readable.json

PROTOS = tester service-tester subdir/docstuff
PROTO_DIR = $(TOP_DIR)/tests/data/proto1
PROTO_FILES = $(foreach proto,$(PROTOS),$(PROTO_DIR)/$(proto).proto)


OUT_DIR = $(TOP_DIR)
BIN_DIR = $(TOP_DIR)/cmd/protoc-gen-docjson

all: plugin

check: plugin
	/usr/bin/env PATH=$(BIN_DIR):$${PATH} protoc \
		--docjson_out="$(OUT_DIR)" \
		--docjson_opt=outfile=$(OUTFILE),proto=$(PROTO_DIR),pretty \
		-I$(PROTO_DIR) \
		$(PROTO_FILES)

checkdebug: plugin
	/usr/bin/env PATH=$(BIN_DIR):$${PATH} protoc \
		--docjson_out="$(OUT_DIR)" \
		--docjson_opt=outfile=$(OUTFILE),proto=$(PROTO_DIR),debug \
		-I$(PROTO_DIR) \
		$(PROTO_FILES)
	cat $(TOP_DIR)/$(OUTFILE) | jq > $(TOP_DIR)/$(READABLE_OUTFILE)

checkyaml: plugin
	/usr/bin/env PATH=$(BIN_DIR):$${PATH} protoc \
		--docjson_out="$(OUT_DIR)" \
		--docjson_opt=proto=$(PROTO_DIR),outfmt=yaml \
		-I$(PROTO_DIR) \
		$(PROTO_FILES)

plugin:
	cd cmd/protoc-gen-docjson && go build -a

test: plugin
	cd tests && go test

testv: plugin
	cd tests && go test -v