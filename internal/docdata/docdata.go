package docdata

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

	"strings"

	desc_pb "google.golang.org/protobuf/types/descriptorpb"
	// Generated code.
	// First-party modules.
)

type Namespace []string

type PluginOpts struct {
	Debug         bool            `json:"debug"`
	DebugSections map[string]bool `json:"debug_sections"`
	Diag          bool            `json:"diag"`
	OutFile       string          `json:"outfile"`
	ProtoPaths    []string        `json:"proto_paths"`
}

type CompilerDiag struct {
	// Formatted version of the protobuf compiler (`protoc`).
	Version string

	// Parameter passed to the plugin. This is the value passed to the protobuf
	// compiler (`protoc`) via the `--docjson_out` parameter.
	PluginParameter string

	// Number of files the plugin is asked to generate.
	NumFiles int
}

type Config struct {
	PluginOpts   *PluginOpts
	CompilerDiag *CompilerDiag
}

type CommentData struct {
	Description             string   `json:"description"`
	LeadingComments         string   `json:"leading_comments"`
	TrailingComments        string   `json:"trailing_comments"`
	LeadingDetachedComments []string `json:"leading_detached_comments"`
}

type FieldData struct {
	CommentData
	TypeName      string                 `json:"type"`
	FullTypeName  string                 `json:"full_type"`
	Kind          string                 `json:"kind"`
	Name          string                 `json:"name"`
	FullName      string                 `json:"full_name"`
	Label         string                 `json:"label"`
	FieldNumber   int32                  `json:"field_number"`
	DefaultValue  string                 `json:"default_value"`
	OneofIndex    int32                  `json:"oneof_index"`
	InOneof       bool                   `json:"in_oneof"`
	OneofName     string                 `json:"oneof_name,omitempty"`
	OneofFullName string                 `json:"oneof_full_name,omitempty"`
	Options       *desc_pb.FieldOptions  `json:"options"`
	CustomOptions map[string]interface{} `json:"custom_options"`

	// File this field was defined in.
	DefinedIn string `json:"defined_in"`
}

type OneOfData struct {
	CommentData
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

type EnumValue struct {
	CommentData
	Name          string
	Number        int32
	Options       *desc_pb.EnumValueOptions `json:"options"`
	CustomOptions []string                  `json:"custom_options"`
}

type EnumData struct {
	CommentData
	Name          string                 `json:"name"`
	FullName      string                 `json:"full_name"`
	Description   string                 `json:"description"`
	Values        []*EnumValue           `json:"values"`
	Options       *desc_pb.EnumOptions   `json:"options"`
	CustomOptions map[string]interface{} `json:"custom_options"`

	// File this enum was defined in.
	DefinedIn string `json:"defined_in"`
}

type MessageData struct {
	CommentData
	Name           string         `json:"name"`
	FullName       string         `json:"full_name"`
	Fields         []*FieldData   `json:"fields"`
	NestedMessages []*MessageData `json:"nested_messages"`
	Enums          []*EnumData    `json:"enums"`
	// ExtensionRanges []*desc_pb.DescriptorProto_ExtensionRange `json:"extension_ranges"`
	OneofDecls    []*OneOfData            `json:"oneof_decl"`
	Options       *desc_pb.MessageOptions `json:"options"`
	CustomOptions map[string]interface{}  `json:"custom_options"`

	// File this message was defined in.
	DefinedIn string `json:"defined_in"`
}

type FileExtension struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	FieldNumber int32  `json:"field_number"`
	Type        string `json:"type"`
	Extendee    string `json:"extendee"`

	// File this extension was defined in.
	DefinedIn string `json:"defined_in"`
}

type SyntaxDecl struct {
	CommentData
	Version string `json:"version"`
}

type MethodData struct {
	CommentData
	Name              string                 `json:"name"`
	FullName          string                 `json:"full_name"`
	RequestType       string                 `json:"request_type"`
	RequestFullType   string                 `json:"request_full_type"`
	RequestStreaming  bool                   `json:"request_streaming"`
	ResponseType      string                 `json:"response_type"`
	ResponseFullType  string                 `json:"response_full_type"`
	ResponseStreaming bool                   `json:"response_streaming"`
	Options           *desc_pb.MethodOptions `json:"options"`
	CustomOptions     map[string]interface{} `json:"custom_options"`

	// File this method was defined in.
	DefinedIn string `json:"defined_in"`
}

type ServiceData struct {
	CommentData
	Name          string                  `json:"name"`
	FullName      string                  `json:"full_name"`
	Methods       []*MethodData           `json:"methods"`
	Options       *desc_pb.ServiceOptions `json:"options"`
	CustomOptions map[string]interface{}  `json:"custom_options"`

	// File this service was defined in.
	DefinedIn string `json:"defined_in"`
}

type FileData struct {
	Name                 string                 `json:"name"`
	Package              string                 `json:"package"`
	Messages             []*MessageData         `json:"messages"`
	Enums                []*EnumData            `json:"enums"`
	Services             []*ServiceData         `json:"services"`
	Dependencies         []string               `json:"dependencies"`
	ExternalDependencies []string               `json:"external_dependencies"`
	Options              *desc_pb.FileOptions   `json:"options"`
	Extensions           []*FileExtension       `json:"extensions"`
	Syntax               *SyntaxDecl            `json:"syntax"`
	CustomOptions        map[string]interface{} `json:"custom_options"`
}

type TemplateData struct {
	// List of protobuf spec file names in the order provided by the protobuf
	// compiler.
	FileList []string `json:"file_name_list"`

	// Map of protobuf spec file name to file details.
	FileMap map[string]*FileData `json:"file_map"`

	// List of fully-qualified service names.
	ServiceList []string `json:"service_name_list"`

	// Map of fully-qualified service names to service details.
	ServiceMap map[string]*ServiceData `json:"service_map"`

	// List of fully-qualified message names.
	MessageList []string `json:"message_name_list"`

	// Map of fully-qualified message names to message details.
	MessageMap map[string]*MessageData `json:"message_map"`

	// List of fully-qualified extension names.
	ExtensionList []string `json:"extension_name_list"`

	// Map of fully-qualified extensions to extension details.
	ExtensionMap map[string]*FileExtension `json:"extension_map"`

	// List of fully-qualified enumeration declaration names.
	EnumList []string `json:"enum_name_list"`

	// Map of fully-qualified enum declaration names to enum details.
	EnumMap map[string]*EnumData `json:"enum_map"`

	// Map of fully-qualified message names to lists of dependent message and
	// enumeration names.
	MessageDeps map[string][]string `json:"message_deps"`

	// Map of fully-qualified service names to lists of dependent message and
	// enumeration names.
	ServiceDeps map[string][]string `json:"service_deps"`
}

func (ns Namespace) QualifyName(name string) string {
	if len(ns) == 0 {
		return name
	}

	return strings.Join(ns, ".") + "." + name
}

func (ns Namespace) Extend(name string) Namespace {
	dst := make(Namespace, len(ns), len(ns)+1)
	copy(dst, ns)
	dst = append(dst, name)

	return dst
}
