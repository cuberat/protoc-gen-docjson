package main

import (
	// Built-in/core modules.
	"bufio"
	json "encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	text_scanner "text/scanner"
	"strings"

	// Third-party modules.
	log "github.com/sirupsen/logrus"
	proto "google.golang.org/protobuf/proto"
	desc_pb "google.golang.org/protobuf/types/descriptorpb"
	// Generated code.
	// First-party modules.
)

func main() {
	var (
		infile               string
		outfile              string
		output_template_data bool
		reader               io.Reader
		writer               io.Writer
	)

	flag.StringVar(&infile, "infile", "", "Input file")
	flag.StringVar(&outfile, "outfile", "", "Output file")
	flag.BoolVar(&output_template_data, "template", false,
		"Output JSON suitable for applying to documentation templates")

	flag.Parse()

	if infile == "" {
		log.Fatal("`infile` parameter required")
	}

	in_fh, err := os.Open(infile)
	if err != nil {
		log.Fatalf("couldn't open %q for input: %s", infile, err)
	}
	reader = in_fh
	defer in_fh.Close()

	if outfile == "" {
		writer = os.Stdout
	} else {
		out_fh, err := os.Create(outfile)
		if err != nil {
			log.Fatalf("couldn't open %q for output: %s", outfile, err)
		}
		writer = out_fh
		defer out_fh.Close()
	}

	if output_template_data {
		err := gen_template_data(reader, writer)
		if err != nil {
			log.Fatalf("couldn't generate template data: %s", err)
		}
		os.Exit(0)
	}

	err = convert_descriptor_set(reader, writer)
	if err != nil {
		log.Fatalf("couldn't convert descriptor set from file %q: %s",
			infile, err)
	}
}

type FieldData struct {
	TypeName string `json:"type"`
	Name string `json:"name"`
	Label string `json:"label"`
	SlotNumber int32 `json:"slot_number"`
	DefaultValue string `json:"default_value"`
	OneofIndex int32 `json:"oneof_index"`
	Options *desc_pb.FieldOptions `json:"options"`
	CustomOptions map[string]interface{} `json:"custom_options"`
	Description string `json:"description"`
}

type EnumData struct {

}

type ServiceData struct {

}

type MessageData struct {
	Name string `json:"name"`
	Fields []*FieldData `json:"fields"`
	NestedMessages []*MessageData `json:"nested_messages"`
	Enums []*EnumData `json:"enums"`
	ExtensionRanges []*desc_pb.DescriptorProto_ExtensionRange `json:"extension_ranges"`
	OneofDecls []*desc_pb.OneofDescriptorProto `json:"oneof_decl"`
	Options *desc_pb.MessageOptions `json:"options"`
	Description string `json:"description"`
}

type FileExtension struct {
	Name string `json:"name"`
	SlotNumber int32 `json:"slot_number"`
	Type string `json:"type"`
	Extendee string `json:"extendee"`
}

type FileData struct {
	Name string `json:"name"`
	Package string `json:"package"`
	Messages []*MessageData `json:"messages"`
	Dependencies []string `json:"dependencies"`
	Options *desc_pb.FileOptions `json:"options"`
	Extensions []*FileExtension `json:"extensions"`
}
type TemplateData struct {
	Files []*FileData `json:"files"`
}

func gen_template_data(reader io.Reader, writer io.Writer) error {
	desc_set, err := unmarshal_descriptor_set(reader)
	if err != nil {
		return err
	}

	template_data := new(TemplateData)
	for _, desc_file_info := range desc_set.File {
		this_file := new(FileData)
		template_data.Files = append(template_data.Files, this_file)

		if desc_file_info.Name != nil {
			this_file.Name = *desc_file_info.Name
		}
		if desc_file_info.Package != nil {
			this_file.Package = *desc_file_info.Package
		}
		this_file.Dependencies = desc_file_info.Dependency
		this_file.Options = desc_file_info.Options

		messages := make([]*MessageData, 0, 0)
		for _, msg := range desc_file_info.MessageType {
			messages = append(messages, get_msg_data_from_desc(msg))
		}

		this_file.Messages = messages

		for _, extension := range desc_file_info.Extension {
			this_extension := new(FileExtension)
			if extension.Name != nil {
				this_extension.Name = *extension.Name
			}
			if extension.Number != nil {
				this_extension.SlotNumber = *extension.Number
			}
			if extension.Type != nil {
				this_extension.Type =
					field_type_enum_to_string(*extension.Type)
			}
			if extension.Extendee != nil {
				this_extension.Extendee = *extension.Extendee
			}

			this_file.Extensions =
				append(this_file.Extensions, this_extension)
		}

		if desc_file_info.SourceCodeInfo != nil {
			add_source_code_info(this_file, desc_file_info.SourceCodeInfo)
		}
	}

	process_extensions(template_data, desc_set)

	json_bytes, err := json.Marshal(template_data)
	if err != nil {
		return fmt.Errorf("couldn't marshal template data to JSON: %s", err)
	}
	writer.Write(json_bytes)

	return nil
}

func process_extensions(
	template_data *TemplateData,
	desc_set *desc_pb.FileDescriptorSet,
) {
	// Collect all of the extensions so that we can resolve custom options as
	// we walk through the structures again.
	extensions := make(map[string]map[int32]*FileExtension)
	for _, file_info := range template_data.Files {
		for _, ext := range file_info.Extensions {
			if ext_type, ok := extensions[ext.Extendee]; ok {
				ext_type[ext.SlotNumber] = ext
			} else {
				extensions[ext.Extendee] = map[int32]*FileExtension{
					ext.SlotNumber: ext,
				}
			}
		}
	}

	for file_index, desc_file_info := range desc_set.File {
		if desc_file_info.SourceCodeInfo == nil {
			continue
		}

		this_file := template_data.Files[file_index]

		for _, loc := range desc_file_info.SourceCodeInfo.Location {
			if len(loc.Path) == 6 {
				// Could be a custom field option.
				loc_path := loc.Path
				if loc_path[0] == 4 {
					// Message
					msg := this_file.Messages[loc_path[1]]

					if loc_path[2] == 2 {
						// Field
						field := msg.Fields[loc_path[3]]
						if loc_path[4] == 8 {
							// Custom field option.
							ext_number := loc_path[5]
							ext_type := ".google.protobuf.FieldOptions"
							extendee, ok := extensions[ext_type]
							if !ok {
								continue
							}
							ext, ok := extendee[ext_number]
							if !ok {
								continue
							}

							span_text := get_text_from_span(this_file.Name,
								loc.Span)
							option_val_str :=
								get_option_val_from_string(span_text)
							// FIXME: convert to correct type
							field.CustomOptions[ext.Name] =
								convert_ext_val(ext, option_val_str)

						}
					}
				}
			}
		}

	}
}

func convert_ext_val(
	ext *FileExtension,
	val_string string,
) interface{} {
	ext_type := ext.Type
	switch ext_type {
	case "bool":
		val_string = strings.ToLower(val_string)
		if val_string == "true" {
			return true
		}
		return false
	}

	return ""
}

func get_option_val_from_string(option_str string) string {
	scanner := new(text_scanner.Scanner)
	scanner.Init(strings.NewReader(option_str))

	scanner.Scan()
	text := scanner.TokenText()
	if text != "(" {
		log.Errorf("failed to parse option string %q: missing '('",
			option_str)
		return ""
	}

	tok := scanner.Scan()
	for ; tok != text_scanner.EOF; tok = scanner.Scan() {
		if scanner.TokenText() == ")" {
			break
		}
	}

	if tok == text_scanner.EOF {
		log.Errorf("reached end of option string %q before parsing complete",
			option_str)
		return ""
	}

	if tok = scanner.Scan(); tok == text_scanner.EOF {
		log.Errorf("reached end of option string %q before parsing complete",
			option_str)
		return ""
	}

	if scanner.TokenText() != "=" {
		log.Errorf("expected '=' instead of %q in option string %s",
			scanner.TokenText(), option_str)
	}

	if tok = scanner.Scan(); tok == text_scanner.EOF {
		log.Errorf("reached end of option string %q before parsing complete",
			option_str)
		return ""
	}

	return scanner.TokenText()
}

func get_text_from_span(file_name string, loc_span []int32) string {
	file_path := path.Join("proto", file_name)
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
		log.Errorf("couldn't find section of source code for span %v " +
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

func add_source_code_info(
	file_data *FileData,
	source_info *desc_pb.SourceCodeInfo,
) {
	for _, location := range source_info.Location {
		comments := ""
		if location.LeadingComments != nil {
			comments += *location.LeadingComments
		}
		if location.TrailingComments != nil {
			comments += *location.TrailingComments
		}

		if comments == "" {
			continue
		}

		loc_path := location.Path
		desc_slot_num := loc_path[0]

		switch desc_slot_num {
		case 4: // message
			msg := file_data.Messages[loc_path[1]]
			add_msg_desc(loc_path, msg, comments)
		case 5: // enum
		case 6: // service
		}

	}
}

func add_msg_desc(loc_path []int32, msg *MessageData, comments string) {
	if len(loc_path) == 2 {
		msg.Description = comments
		return
	}
	loc_path = loc_path[2:]

	slot_num := loc_path[0]
	switch slot_num {
	case 2:
		// Field comment.
		field := msg.Fields[loc_path[1]]
		field.Description = comments
	}
}

func get_msg_data_from_desc(msg *desc_pb.DescriptorProto) *MessageData {
	this_msg := new(MessageData)
	if msg.Name != nil {
		this_msg.Name = *msg.Name
	}

	fields := make([]*FieldData, 0, 0)
	for _, field_info := range msg.Field {
		fields = append(fields, get_field_data_from_desc(field_info))
	}
	this_msg.Fields = fields

	this_msg.ExtensionRanges = msg.ExtensionRange
	this_msg.OneofDecls = msg.OneofDecl
	this_msg.Options = msg.Options

	return this_msg
}

func get_field_data_from_desc(
	field *desc_pb.FieldDescriptorProto,
) *FieldData {
	this_field := new(FieldData)

	if field.Name != nil {
		this_field.Name = *field.Name
	}
	if field.Number != nil {
		this_field.SlotNumber = *field.Number
	}
	if field.Label != nil {
		s := desc_pb.FieldDescriptorProto_Label_name[int32(*field.Label)]
		s = strings.ToLower(s)
		this_field.Label = strings.TrimPrefix(s, "label_")
	}
	if field.TypeName != nil {
		this_field.TypeName = *field.TypeName
	} else if (field.Type != nil){
		this_field.TypeName = field_type_enum_to_string(*field.Type)
	}

	if field.DefaultValue != nil {
		this_field.DefaultValue = *field.DefaultValue
	}

	if field.OneofIndex != nil {
		this_field.OneofIndex = *field.OneofIndex
	}

	if field.Options != nil {
		this_field.Options = field.Options
	}

	this_field.CustomOptions = make(map[string]interface{}, 0)

	return this_field
}

func field_type_enum_to_string(
	field_type desc_pb.FieldDescriptorProto_Type,
) string {
	enum_str := desc_pb.FieldDescriptorProto_Type_name[int32(field_type)]
	enum_str = strings.ToLower(enum_str)
	return strings.TrimPrefix(enum_str, "type_")
}

func convert_descriptor_set(reader io.Reader, writer io.Writer) error {
	desc_set, err := unmarshal_descriptor_set(reader)
	if err != nil {
		return err
	}

	json_bytes, err := json.Marshal(desc_set)
	if err != nil {
		return fmt.Errorf("couldn't marshal FileDescriptorSet to JSON: %w",
			err)
	}

	_, err = writer.Write(json_bytes)
	if err != nil {
		return fmt.Errorf("couldn't write out JSON: %w", err)
	}

	return nil
}

func unmarshal_descriptor_set(
	reader io.Reader,
) (*desc_pb.FileDescriptorSet, error) {
	raw_pb, err := io.ReadAll(reader)
	if err != nil {
		return nil,
			fmt.Errorf("couldn't read from file: %w", err)
	}

	desc_set := new(desc_pb.FileDescriptorSet)
	err = proto.Unmarshal(raw_pb, desc_set)
	if err != nil {
		return nil,
			fmt.Errorf("couldn't unmarshal to FileDescriptorSet: %w", err)
	}

	return desc_set, nil
}
