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
	OutFile       string          `json:"out_file"`
	OutFormat     string          `json:"out_format"`
	ProtoPaths    []string        `json:"proto_paths"`
	PrettyPrint   bool            `json:"pretty_out"`
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

type FieldOptions struct {
	CType      desc_pb.FieldOptions_CType  `json:"ctype"`
	Packed     bool                        `json:"packed"`
	JSType     desc_pb.FieldOptions_JSType `json:"jstype"`
	Lazy       bool                        `json:"lazy"`
	Deprecated bool                        `json:"deprecated"`
}

type FieldData struct {
	CommentData
	TypeName      string         `json:"type"`
	FullTypeName  string         `json:"full_type"`
	Kind          string         `json:"kind"`
	Name          string         `json:"name"`
	FullName      string         `json:"full_name"`
	Label         string         `json:"label"`
	FieldNumber   int32          `json:"field_number"`
	DefaultValue  string         `json:"default_value"`
	OneofIndex    int32          `json:"oneof_index"`
	InOneof       bool           `json:"in_oneof"`
	OneofName     string         `json:"oneof_name"`
	OneofFullName string         `json:"oneof_full_name"`
	Options       *FieldOptions  `json:"options"`
	CustomOptions map[string]any `json:"custom_options"`

	// File this field was defined in.
	DefinedIn string `json:"defined_in"`
}

type OneOfData struct {
	CommentData
	Name     string `json:"name"`
	FullName string `json:"full_name"`

	// The only option for a oneof is the `uninterpreted_option` used to
	// temporarily hold data for the parser. So there is no `Options`
	// field here.
}

type EnumValueOptions struct {
	Deprecated bool `json:"deprecated"`
}

type EnumValue struct {
	CommentData
	Name          string            `json:"name"`
	Number        int32             `json:"number"`
	Options       *EnumValueOptions `json:"options"`
	CustomOptions map[string]any    `json:"custom_options"`
}

type EnumOptions struct {
	AllowAlias bool `json:"allow_alias"`
	Deprecated bool `json:"deprecated"`
}

type EnumData struct {
	CommentData
	Name          string         `json:"name"`
	FullName      string         `json:"full_name"`
	Description   string         `json:"description"`
	Values        []*EnumValue   `json:"values"`
	Options       *EnumOptions   `json:"options"`
	CustomOptions map[string]any `json:"custom_options"`

	// File this enum was defined in.
	DefinedIn string `json:"defined_in"`
}

type MessageOptions struct {
	Deprecated bool `json:"deprecated"`
	MapEntry   bool `json:"map_entry"`
}

type MessageData struct {
	CommentData
	Name           string          `json:"name"`
	FullName       string          `json:"full_name"`
	Fields         []*FieldData    `json:"fields"`
	NestedMessages []*MessageData  `json:"nested_messages"`
	Enums          []*EnumData     `json:"enums"`
	OneofDecls     []*OneOfData    `json:"oneof_decl"`
	Options        *MessageOptions `json:"options"`
	CustomOptions  map[string]any  `json:"custom_options"`

	// File this message was defined in.
	DefinedIn string `json:"defined_in"`
}

type FileExtension struct {
	CommentData
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

type MethodOptions struct {
	Deprecated bool `json:"deprecated"`
}

type MethodData struct {
	CommentData
	Name              string         `json:"name"`
	FullName          string         `json:"full_name"`
	RequestType       string         `json:"request_type"`
	RequestFullType   string         `json:"request_full_type"`
	RequestStreaming  bool           `json:"request_streaming"`
	ResponseType      string         `json:"response_type"`
	ResponseFullType  string         `json:"response_full_type"`
	ResponseStreaming bool           `json:"response_streaming"`
	Options           *MethodOptions `json:"options"`
	CustomOptions     map[string]any `json:"custom_options"`

	// File this method was defined in.
	DefinedIn string `json:"defined_in"`
}

type ServiceOptions struct {
	Deprecated bool `json:"deprecated"`
}

type ServiceData struct {
	CommentData
	Name          string          `json:"name"`
	FullName      string          `json:"full_name"`
	Methods       []*MethodData   `json:"methods"`
	Options       *ServiceOptions `json:"options"`
	CustomOptions map[string]any  `json:"custom_options"`

	// File this service was defined in.
	DefinedIn string `json:"defined_in"`
}

type FileOptions struct {
	JavaPackage          string `json:"java_package"`
	JavaOuterClassname   string `json:"java_outer_classname"`
	JavaMultipleFiles    bool   `json:"java_multiple_files"`
	JavaStringCheckUtf8  bool   `json:"java_string_check_utf8"`
	GoPackage            string `json:"go_package"`
	Deprecated           bool   `json:"deprecated"`
	CcEnableArenas       bool   `json:"cc_enable_arenas"`
	ObjcClassPrefix      string `json:"objc_class_prefix"`
	CsharpNamespace      string `json:"csharp_namespace"`
	SwiftPrefix          string `json:"swift_prefix"`
	PhpClassPrefix       string `json:"php_class_prefix"`
	PhpNamespace         string `json:"php_namespace"`
	PhpMetadataNamespace string `json:"php_metadata_namespace"`
	RubyPackage          string `json:"ruby_package"`
}

type FileData struct {
	Name                 string           `json:"name"`
	Package              string           `json:"package"`
	Messages             []*MessageData   `json:"messages"`
	Enums                []*EnumData      `json:"enums"`
	Services             []*ServiceData   `json:"services"`
	Dependencies         []string         `json:"dependencies"`
	ExternalDependencies []string         `json:"external_dependencies"`
	Options              *FileOptions     `json:"options"`
	Extensions           []*FileExtension `json:"extensions"`
	Syntax               *SyntaxDecl      `json:"syntax"`
	CustomOptions        map[string]any   `json:"custom_options"`

	// File extensions that extend protobuf option messages.
	DeclaredCustomOptions map[string][]*FileExtension `json:"declared_custom_options"`
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

	// Map of fully-qualified service names to lists of dependent files.
	ServiceFileDeps map[string][]string `json:"service_file_deps"`
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
