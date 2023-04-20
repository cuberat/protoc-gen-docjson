package plugin

// This file contains the code to ingest data from the protobuf compiler
// (`protoc`) and respond with the appropriate file content, which the
// protobuf compiler then writes out. See
// https://pkg.go.dev/google.golang.org/protobuf/types/pluginpb for details.

// BSD 2-Clause License
//
// Copyright (c) 2023 Don Owens <don@regexguy.com>.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice,
//   this list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

import (
	// Built-in/core modules.
	json "encoding/json"
	"fmt"
	"io"
	"strings"

	// Third-party modules.

	log "github.com/sirupsen/logrus"
	proto "google.golang.org/protobuf/proto"
	desc_pb "google.golang.org/protobuf/types/descriptorpb"
	pluginpb "google.golang.org/protobuf/types/pluginpb"
	yaml "gopkg.in/yaml.v3"

	// Generated code.
	// First-party modules.
	docdata "github.com/cuberat/protoc-gen-docjson/internal/docdata"
	docgen "github.com/cuberat/protoc-gen-docjson/internal/docgen"
)

func ProcessCodeGenRequest(
	reader io.Reader,
	writer io.Writer,
) error {
	gen_req := new(pluginpb.CodeGeneratorRequest)

	if raw_request, err := io.ReadAll(reader); err != nil {
		return fmt.Errorf("read failed: %w", err)
	} else {
		err = proto.Unmarshal(raw_request, gen_req)
		if err != nil {
			return fmt.Errorf("couldn't unmarshal CodeGenerationRequest: %w",
				err)
		}
	}

	conf := setup_config(gen_req)

	files_to_generate := map[string]bool{}
	for _, file_name := range gen_req.FileToGenerate {
		files_to_generate[file_name] = true
	}

	protos_to_process := make([]*desc_pb.FileDescriptorProto, 0, 1)
	for _, file_desc := range gen_req.ProtoFile {
		if files_to_generate[file_desc.GetName()] {
			protos_to_process = append(protos_to_process, file_desc)
		}
	}

	template_data, err := docgen.GenDocData(conf, protos_to_process,
		files_to_generate)
	if err != nil {
		err = fmt.Errorf("couldn't generate template data: %w", err)
		send_code_gen_err(err, writer)
		return err
	}

	content, err := serialize_content(template_data, conf)
	if err != nil {
		return send_code_gen_err(err, writer)
	}
	file := &pluginpb.CodeGeneratorResponse_File{
		Name:    &conf.PluginOpts.OutFile,
		Content: &content,
	}

	gen_resp := new(pluginpb.CodeGeneratorResponse)
	gen_resp.File = []*pluginpb.CodeGeneratorResponse_File{file}

	return send_code_gen_resp(gen_resp, writer)
}

func serialize_content(
	data *docdata.TemplateData,
	conf *docdata.Config,
) (string, error) {
	out_format := conf.PluginOpts.OutFormat
	out_file := conf.PluginOpts.OutFile
	if out_format == "" {
		switch {
		case strings.HasSuffix(out_file, ".yaml"):
			fallthrough
		case strings.HasSuffix(out_file, ".yml"):
			out_format = "yaml"
		case strings.HasSuffix(out_file, ".json"):
			out_format = "json"
		default:
			out_format = "json"
		}
	}

	defer func() {
		if conf.PluginOpts.OutFile == "" {
			conf.PluginOpts.OutFile = "doc." + out_format
		}
	}()

	switch out_format {
	case "yaml":
		return marshal_to_yaml(data)
	default:
		return marshal_to_json(data)
	}
}

func marshal_to_json(data any) (string, error) {
	buffer_bytes, err := json.Marshal(data)
	// marshaler := &protojson.MarshalOptions{
	// 	Multiline:       false,
	// 	Indent:          "",
	// 	AllowPartial:    true,
	// 	UseProtoNames:   true,
	// 	UseEnumNumbers:  false,
	// 	EmitUnpopulated: true,
	// 	// Resolver: nil,
	// }
	// buffer_bytes, err := marshaler.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("couldn't marshal to JSON: %w", err)
	}

	return string(buffer_bytes), nil
}

func marshal_to_yaml(data any) (string, error) {
	buffer_bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("couldn't marshal to YAML: %w", err)
	}

	return string(buffer_bytes), nil
}

func setup_config(gen_req *pluginpb.CodeGeneratorRequest) *docdata.Config {
	plugin_opts := parse_plugin_option(gen_req.GetParameter())

	conf := &docdata.Config{
		PluginOpts: plugin_opts,
	}

	if conf.PluginOpts.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	populate_diag_info(gen_req, conf)

	if plugin_opts.Diag {
		log.WithFields(log.Fields{
			"compiler version": conf.CompilerDiag.Version,
			"plugin parameter": conf.CompilerDiag.PluginParameter,
			"number of files":  conf.CompilerDiag.NumFiles,
		}).Info("Compiler information:")
	}

	return conf
}

func populate_diag_info(
	gen_req *pluginpb.CodeGeneratorRequest,
	conf *docdata.Config,
) {
	version_info := gen_req.GetCompilerVersion()
	version := fmt.Sprintf("%d.%d", version_info.GetMajor(),
		version_info.GetMinor())
	if version_info.Patch != nil {
		version += fmt.Sprintf(".%d", version_info.GetPatch())
	}
	if version_info.Suffix != nil {
		version += fmt.Sprintf(".%s", version_info.GetSuffix())
	}

	conf.CompilerDiag = &docdata.CompilerDiag{
		Version:         version,
		PluginParameter: gen_req.GetParameter(),
		NumFiles:        len(gen_req.GetFileToGenerate()),
	}
}

func parse_plugin_option(opts string) *docdata.PluginOpts {
	options := &docdata.PluginOpts{
		DebugSections: make(map[string]bool),
	}
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
		case "diag":
			options.Diag = true
		case "debug":
			options.DebugSections[opt_pair[1]] = true
		case "outfmt":
			options.OutFormat = opt_pair[1]
		}
	}

	log.Debugf("plugin options: %+v", options)

	return options
}

func send_code_gen_err(err error, writer io.Writer) error {
	gen_resp := new(pluginpb.CodeGeneratorResponse)
	err_str := err.Error()
	gen_resp.Error = &err_str
	return send_code_gen_resp(gen_resp, writer)
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
