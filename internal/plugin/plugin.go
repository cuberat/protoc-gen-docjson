package plugin

import (
	// Built-in/core modules.

	// Third-party modules.
	"encoding/json"
	"fmt"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
	proto "google.golang.org/protobuf/proto"
	desc_pb "google.golang.org/protobuf/types/descriptorpb"
	pluginpb "google.golang.org/protobuf/types/pluginpb"

	// Generated code.
	// First-party modules.
	docdata "github.com/cuberat/protoc-gen-docjson/internal/docdata"
	docgen "github.com/cuberat/protoc-gen-docjson/internal/docgen"
)

func ProcessCodeGenRequest(
	reader io.Reader,
	writer io.Writer,
) error {
	raw_request, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read failed: %w", err)
	}

	gen_req := new(pluginpb.CodeGeneratorRequest)
	err = proto.Unmarshal(raw_request, gen_req)
	if err != nil {
		return fmt.Errorf("couldn't unmarshal CodeGenerationRequest: %w",
			err)
	}

	plugin_opts := parse_plugin_option(gen_req.GetParameter())
	if plugin_opts.OutFile == "" {
		plugin_opts.OutFile = "doc.json"
	}

	if plugin_opts.Debug {
		log.SetLevel(log.DebugLevel)
		// log.SetReportCaller(true)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	files_to_generate := map[string]bool{}
	for _, file_name := range gen_req.FileToGenerate {
		files_to_generate[file_name] = true
	}

	log.Debugf("file_to_generate: %s", gen_req.FileToGenerate)
	log.Debugf("parameter: %s", gen_req.GetParameter())

	log.Debugf("compiler_version: %s", gen_req.GetCompilerVersion())

	protos_to_process := make([]*desc_pb.FileDescriptorProto, 0, 1)

	for _, file_desc := range gen_req.ProtoFile {
		name := file_desc.GetName()
		pkg := file_desc.GetPackage()

		log.Debugf("file_desc %s - %s", name, pkg)

		if files_to_generate[name] {
			protos_to_process = append(protos_to_process, file_desc)
		}
	}

	for _, file_desc := range protos_to_process {
		log.Debugf("---> will process %q", *file_desc.Name)
	}

	gen_resp := new(pluginpb.CodeGeneratorResponse)

	template_data, err := docgen.GenDocData(plugin_opts, protos_to_process,
		files_to_generate)
	if err != nil {
		err = fmt.Errorf("couldn't generate template data: %w", err)
		send_code_gen_err(err, writer)
		return err
	}

	file := new(pluginpb.CodeGeneratorResponse_File)

	file.Name = &plugin_opts.OutFile

	json_bytes, err := json.Marshal(template_data)
	if err != nil {
		err = fmt.Errorf("couldn't marshal template data to JSON: %s", err)
		send_code_gen_err(err, writer)
		return err
	}

	content := string(json_bytes)
	file.Content = &content

	gen_resp.File = []*pluginpb.CodeGeneratorResponse_File{file}

	send_code_gen_resp(gen_resp, writer)

	return nil
}

func parse_plugin_option(opts string) *docdata.PluginOpts {
	options := new(docdata.PluginOpts)
	if opts == "" {
		return options
	}

	opts_list := strings.Split(opts, ",")
	for _, opt := range opts_list {
		if opt == "debug" {
			options.Debug = true
			continue
		}

		opt_pair := strings.SplitN(opt, "=", 2)
		switch strings.TrimSpace(opt_pair[0]) {
		case "outfile":
			options.OutFile = strings.TrimSpace(opt_pair[1])
		case "proto":
			options.ProtoPaths =
				append(options.ProtoPaths, strings.TrimSpace(opt_pair[1]))
		}
	}

	log.Debugf("plugin options: %+v", options)

	return options
}

func send_code_gen_err(err error, writer io.Writer) {
	gen_resp := new(pluginpb.CodeGeneratorResponse)
	err_str := err.Error()
	gen_resp.Error = &err_str
	send_code_gen_resp(gen_resp, writer)
}

func send_code_gen_resp(
	resp *pluginpb.CodeGeneratorResponse,
	writer io.Writer,
) error {
	raw_resp, err := proto.Marshal(resp)
	if err != nil {
		return fmt.Errorf("couldn't marshal CodeGeneratorResponse: %w", err)
	}

	if _, err = writer.Write(raw_resp); err != nil {
		return fmt.Errorf("couldn't write response to compiler: %w", err)
	}

	return nil
}
