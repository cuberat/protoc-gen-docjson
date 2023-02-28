package extensions

import (
	// Built-in/core modules.
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	// Third-party modules.
	textparser "github.com/cuberat/go-textparser"
	log "github.com/sirupsen/logrus"
	desc_pb "google.golang.org/protobuf/types/descriptorpb"

	// Generated code.
	// First-party modules.
	docdata "github.com/cuberat/protoc-gen-docjson/internal/docdata"
)

type CustomOptionProcessor struct {
	Extensions map[string]map[int32]*docdata.FileExtension
	File       *docdata.FileData
	PluginOpts *docdata.PluginOpts
}

func ProcessExtensions(
	template_data *docdata.TemplateData,
	file_descriptors []*desc_pb.FileDescriptorProto,
	plugin_opts *docdata.PluginOpts,
) {
	// Collect all of the extensions so that we can resolve custom options as
	// we walk through the structures again.
	extensions := make(map[string]map[int32]*docdata.FileExtension)
	for _, file_info := range template_data.FileMap {
		for _, ext := range file_info.Extensions {
			if ext_type, ok := extensions[ext.Extendee]; ok {
				ext_type[ext.FieldNumber] = ext
			} else {
				extensions[ext.Extendee] = map[int32]*docdata.FileExtension{
					ext.FieldNumber: ext,
				}
			}
		}
	}

	for _, desc_file_info := range file_descriptors {
		if desc_file_info.SourceCodeInfo == nil {
			continue
		}

		this_file := template_data.FileMap[desc_file_info.GetName()]

		opt_processor := &CustomOptionProcessor{
			Extensions: extensions,
			File:       this_file,
			PluginOpts: plugin_opts,
		}

		for _, loc := range desc_file_info.SourceCodeInfo.Location {
			opt_processor.ExtractFileOptions(this_file, loc.Path, loc)
		}

	}
}

func (proc *CustomOptionProcessor) ExtractFieldOptions(
	field *docdata.FieldData,
	loc_path []int32,
	loc *desc_pb.SourceCodeInfo_Location,
) {
	if loc_path[0] == 8 && len(loc_path) == 2 {
		// Custom field option.
		name, val, err := proc.BuildOptionVal(loc_path[1],
			".google.protobuf.FieldOptions", loc)
		if err != nil {
			return
		}

		if field.CustomOptions == nil {
			field.CustomOptions = make(map[string]interface{})
		}

		field.CustomOptions[name] = val

		log.Debugf("found custom field option %q = %v", name, val)

		return
	}
}

func (proc *CustomOptionProcessor) ExtractServiceOptions(
	svc *docdata.ServiceData,
	loc_path []int32,
	loc *desc_pb.SourceCodeInfo_Location,
) {
	if loc_path[0] == 3 && len(loc_path) == 2 {
		// Custom service option.

		name, val, err := proc.BuildOptionVal(loc_path[1],
			".google.protobuf.ServiceOptions", loc)
		if err != nil {
			return
		}

		if svc.CustomOptions == nil {
			svc.CustomOptions = make(map[string]interface{})
		}

		svc.CustomOptions[name] = val

		log.Debugf("found custom service option %q = %v", name, val)

		return
	}

	if loc_path[0] == 2 && len(loc_path) > 2 {
		// Potential custom method option.

		method := svc.Methods[loc_path[1]]
		proc.ExtractMethodOptions(method, loc_path[2:], loc)
	}
}

func (proc *CustomOptionProcessor) ExtractMethodOptions(
	method *docdata.MethodData,
	loc_path []int32,
	loc *desc_pb.SourceCodeInfo_Location,
) {
	if loc_path[0] == 4 && len(loc_path) == 2 {
		// Custom metho option.
		name, val, err := proc.BuildOptionVal(loc_path[1],
			".google.protobuf.MethodOptions", loc)
		if err != nil {
			return
		}

		if method.CustomOptions == nil {
			method.CustomOptions = make(map[string]interface{})
		}

		method.CustomOptions[name] = val

		log.Debugf("found custom method option %q = %v", name, val)

		return
	}
}

func (proc *CustomOptionProcessor) ExtractMessageOptions(
	msg *docdata.MessageData,
	loc_path []int32,
	loc *desc_pb.SourceCodeInfo_Location,
) {
	if loc_path[0] == 2 && len(loc_path) > 2 {
		// Field
		field := msg.Fields[loc_path[1]]
		proc.ExtractFieldOptions(field, loc_path[2:], loc)
		return
	}

	if loc_path[0] == 3 && len(loc_path) > 2 {
		// Nested message.
		field := msg.NestedMessages[loc_path[1]]
		proc.ExtractMessageOptions(field, loc_path[2:], loc)
		return
	}

	if loc_path[0] == 4 && len(loc_path) > 2 {
		enum := msg.Enums[loc_path[1]]
		proc.ExtractEnumOptions(enum, loc_path[2:], loc)
		return
	}

	if loc_path[0] == 7 && len(loc_path) == 2 {
		name, val, err := proc.BuildOptionVal(loc_path[1],
			".google.protobuf.MessageOptions", loc)
		if err != nil {
			return
		}

		if msg.CustomOptions == nil {
			msg.CustomOptions = make(map[string]interface{})
		}

		msg.CustomOptions[name] = val

		log.Debugf("found custom message option %q = %v", name, val)

		return
	}
}

func (proc *CustomOptionProcessor) ExtractEnumOptions(
	enum_data *docdata.EnumData,
	loc_path []int32,
	loc *desc_pb.SourceCodeInfo_Location,
) {
	if len(loc_path) == 0 {
		return
	}

	if loc_path[0] == 3 && len(loc_path) == 2 {
		if enum_data.CustomOptions == nil {
			enum_data.CustomOptions = make(map[string]interface{})
		}

		name, val, err := proc.BuildOptionVal(loc_path[1],
			".google.protobuf.EnumOptions", loc)
		if err != nil {
			return
		}

		if enum_data.CustomOptions == nil {
			enum_data.CustomOptions = make(map[string]interface{})
		}

		enum_data.CustomOptions[name] = val

		log.Debugf("found custom enum option %q = %v", name, val)

		return
	}
}

func (proc *CustomOptionProcessor) ExtractFileOptions(
	this_file *docdata.FileData,
	loc_path []int32,
	loc *desc_pb.SourceCodeInfo_Location,
) {
	if len(loc_path) == 0 {
		return
	}

	if loc_path[0] == 4 && len(loc_path) > 2 {
		// Message
		msg := this_file.Messages[loc_path[1]]
		proc.ExtractMessageOptions(msg, loc_path[2:], loc)
		return
	}

	if loc_path[0] == 5 && len(loc_path) > 2 {
		enum_data := this_file.Enums[loc_path[1]]
		proc.ExtractEnumOptions(enum_data, loc_path[2:], loc)
		return
	}

	if loc_path[0] == 6 && len(loc_path) > 2 {
		// Service
		svc := this_file.Services[loc_path[1]]
		proc.ExtractServiceOptions(svc, loc_path[2:], loc)
		return
	}

	if loc_path[0] == 8 && len(loc_path) == 2 {
		name, val, err := proc.BuildOptionVal(loc_path[1],
			".google.protobuf.FileOptions", loc)
		if err != nil {
			return
		}

		if this_file.CustomOptions == nil {
			this_file.CustomOptions = make(map[string]interface{})
		}

		this_file.CustomOptions[name] = val

		log.Debugf("found custom file option %q = %v", name, val)

		return
	}
}

func (proc *CustomOptionProcessor) BuildOptionVal(
	ext_number int32,
	ext_type string,
	loc *desc_pb.SourceCodeInfo_Location,
) (string, interface{}, error) {
	extendee, ok := proc.Extensions[ext_type]
	if !ok {
		return "", nil, fmt.Errorf("no such extension type %q", ext_type)
	}
	ext, ok := extendee[ext_number]
	if !ok {
		return "", nil, fmt.Errorf("no such extendee number %d", ext_number)
	}

	span_text := get_text_from_span(proc.File.Name, loc.Span, proc.PluginOpts)
	if span_text == "" {
		return "", nil,
			fmt.Errorf("couldn't get span text for custom option %q", ext.Name)
	}

	option_val_str := get_option_val_from_string(span_text)

	return ext.Name, convert_ext_val(ext, option_val_str), nil
}

func convert_ext_val(
	ext *docdata.FileExtension,
	val_string string,
) interface{} {
	ext_type := ext.Type
	switch ext_type {
	case "double", "float":
		flt, err := strconv.ParseFloat(val_string, 64)
		if err != nil {
			log.Errorf("unable to parse %s value %q: %s", ext_type, val_string,
				err)
		}
		return flt

	case "int64", "fixed64", "sfixed64", "sint64":
		num, err := strconv.ParseInt(val_string, 0, 64)
		if err != nil {
			log.Errorf("unable to parse %s value %q", ext_type, val_string)
		}
		return num

	case "uint64":
		num, err := strconv.ParseUint(val_string, 0, 64)
		if err != nil {
			log.Errorf("unable to parse %s value %q", ext_type, val_string)
		}
		return num

	case "int32", "fixed32", "sfixed32", "sint32":
		num, err := strconv.ParseInt(val_string, 0, 32)
		if err != nil {
			log.Errorf("unable to parse %s value %q", ext_type, val_string)
		}
		return int32(num)

	case "bool":
		val_string = strings.ToLower(val_string)
		if val_string == "true" {
			return true
		}
		return false

	case "string":
		return strings.TrimSuffix(strings.TrimPrefix(val_string, "\""), "\"")

	case "bytes":
		return []byte(strings.TrimSuffix(strings.TrimPrefix(val_string, "\""), "\""))

	case "uint32":
		num, err := strconv.ParseUint(val_string, 0, 32)
		if err != nil {
			log.Errorf("unable to parse %s value %q", ext_type, val_string)
		}
		return uint32(num)
	}

	return ""
}

func get_option_val_from_string(option_str string) string {
	scanner := textparser.NewScannerString(option_str)

	scanner.Scan()
	text := scanner.TokenText()
	log.Debugf("get_option_val_from_string: got text %q", text)
	if text == "option" {
		// Skip to next token.
		scanner.Scan()
		text = scanner.TokenText()
		log.Debugf("get_option_val_from_string: got text %q", text)
	}

	if text != "(" {
		log.Errorf("failed to parse option string %q: missing '('",
			option_str)
		return ""
	}

	if !scanner.Scan() {
		log.Errorf("reached end of option string %q before parsing complete",
			option_str)
		return ""
	}

	option_name := scanner.TokenText()

	for scanner.Scan() {
		if scanner.TokenText() == ")" {
			break
		}
		option_name += scanner.TokenText()
	}

	log.Debugf("found option name %q", option_name)

	if scanner.TokenText() != ")" {
		log.Errorf("reached end of option string %q before parsing complete:"+
			"missing ')'", option_str)
		return ""
	}

	if !scanner.Scan() || scanner.TokenText() != "=" {
		log.Errorf("reached end of option string %q before parsing complete:"+
			"missing '='", option_str)
	}

	if !scanner.Scan() {
		log.Errorf("reached end of option string %q before parsing complete:"+
			"missing value", option_str)
	}

	return scanner.TokenText()
}

// func hide_get_option_val_from_string(option_str string) string {
// 	scanner := new(text_scanner.Scanner)
// 	scanner.Init(strings.NewReader(option_str))

// 	scanner.Scan()
// 	text := scanner.TokenText()
// 	if text == "option" {
// 		// Skip to next token.
// 		scanner.Scan()
// 		text = scanner.TokenText()
// 	}
// 	if text != "(" {
// 		log.Errorf("failed to parse option string %q: missing '('",
// 			option_str)
// 		return ""
// 	}

// 	tok := scanner.Scan()
// 	for ; tok != text_scanner.EOF; tok = scanner.Scan() {
// 		if scanner.TokenText() == ")" {
// 			break
// 		}
// 	}

// 	if tok == text_scanner.EOF {
// 		log.Errorf("reached end of option string %q before parsing complete",
// 			option_str)
// 		return ""
// 	}

// 	if tok = scanner.Scan(); tok == text_scanner.EOF {
// 		log.Errorf("reached end of option string %q before parsing complete",
// 			option_str)
// 		return ""
// 	}

// 	if scanner.TokenText() != "=" {
// 		log.Errorf("expected '=' instead of %q in option string %s",
// 			scanner.TokenText(), option_str)
// 	}

// 	if tok = scanner.Scan(); tok == text_scanner.EOF {
// 		log.Errorf("reached end of option string %q before parsing complete",
// 			option_str)
// 		return ""
// 	}

// 	if scanner.TokenText() == "-" {
// 		if tok = scanner.Scan(); tok == text_scanner.EOF {
// 			log.Errorf("reached end of option string %q before parsing complete",
// 				option_str)
// 			return ""
// 		}
// 		return "-" + scanner.TokenText()
// 	}

// 	return scanner.TokenText()
// }

func find_file_in_paths(paths []string, file_name string) (string, error) {
	my_paths := make([]string, len(paths), len(paths)+1)
	copy(my_paths, paths)
	my_paths = append(my_paths, ".")

	for _, dir := range my_paths {
		file_path := path.Join(dir, file_name)
		_, err := os.Stat(file_path)
		if err == nil {
			return file_path, nil
		}
	}
	return "", fmt.Errorf("couldn't find proto file %q in paths %v",
		file_name, paths)
}

func get_text_from_span(file_name string,
	loc_span []int32,
	plugin_opts *docdata.PluginOpts,
) string {
	file_path, err := find_file_in_paths(plugin_opts.ProtoPaths, file_name)
	if err != nil {
		log.Error(err)
		return ""
	}

	in_fh, err := os.Open(file_path)
	if err != nil {
		log.Errorf("couldn't open input file %s: %s", file_path, err)
		return ""
	}
	defer in_fh.Close()

	scanner := bufio.NewScanner(in_fh)
	start_line := loc_span[0]
	end_line := start_line
	start_col := loc_span[1]
	end_col := loc_span[2]
	if len(loc_span) == 4 {
		end_line = loc_span[2]
		end_col = loc_span[3]
	}

	line_num := int32(-1)
	for scanner.Scan() {
		line_num++
		if line_num == start_line {
			break
		}
	}

	if line_num != start_line {
		// Short read.
		log.Errorf("couldn't find section of source code for span %v "+
			"in %s", loc_span, file_path)
		return ""
	}

	text := ""

	line := scanner.Text()

	if end_line == start_line {
		text = line[:end_col]
	} else {
		for scanner.Scan() {
			line := scanner.Text()
			line_num++
			if line_num == end_line {
				text += line[:end_col]
				break
			} else {
				text += line
			}
		}
	}

	return text[start_col:]
}
