package extensions

import (
	// Built-in/core modules.
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	text_scanner "text/scanner"

	// Third-party modules.
	log "github.com/sirupsen/logrus"
	desc_pb "google.golang.org/protobuf/types/descriptorpb"

	// Generated code.
	// First-party modules.
	docdata "github.com/cuberat/protoc-gen-docjson/internal/docdata"
)

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
								loc.Span, plugin_opts)
							if span_text == "" {
								continue
							}
							option_val_str :=
								get_option_val_from_string(span_text)

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
	ext *docdata.FileExtension,
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
