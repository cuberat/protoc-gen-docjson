package docgen

import (
	// Built-in/core modules.

	"strings"

	// Third-party modules.
	log "github.com/sirupsen/logrus"
	desc_pb "google.golang.org/protobuf/types/descriptorpb"

	// Generated code.
	// First-party modules.
	docdata "github.com/cuberat/protoc-gen-docjson/internal/docdata"
	extensions "github.com/cuberat/protoc-gen-docjson/internal/extensions"
)

func GenDocData(
	plugin_opts *docdata.PluginOpts,
	file_descriptors []*desc_pb.FileDescriptorProto,
	files_to_generate map[string]bool,
) (*docdata.TemplateData, error) {
	template_data := &docdata.TemplateData{
		FileList: make([]string, 0, len(file_descriptors)),
		FileMap:  make(map[string]*docdata.FileData, len(file_descriptors)),
	}

	for _, desc_file_info := range file_descriptors {
		this_file := new(docdata.FileData)
		// template_data.Files = append(template_data.Files, this_file)

		this_file.Name = desc_file_info.GetName()

		template_data.FileList = append(template_data.FileList, this_file.Name)
		template_data.FileMap[this_file.Name] = this_file

		this_file.Package = desc_file_info.GetPackage()

		this_file.Dependencies =
			make([]string, 0, len(desc_file_info.Dependency))
		this_file.ExternalDependencies =
			make([]string, 0, len(desc_file_info.Dependency))
		for _, file := range desc_file_info.Dependency {
			if files_to_generate[file] {
				this_file.Dependencies = append(this_file.Dependencies, file)
			} else {
				this_file.ExternalDependencies =
					append(this_file.ExternalDependencies, file)
			}
		}
		this_file.Options = desc_file_info.Options

		this_file.Syntax = new(docdata.SyntaxDecl)
		if desc_file_info.Syntax != nil {
			this_file.Syntax.Version = *desc_file_info.Syntax
		}

		messages := make([]*docdata.MessageData, 0, len(desc_file_info.MessageType))
		for _, msg := range desc_file_info.MessageType {
			messages = append(messages,
				get_msg_data_from_desc(msg,
					docdata.Namespace{this_file.Package}))
		}

		this_file.Messages = messages
		this_file.Enums = get_enum_data(desc_file_info.EnumType,
			docdata.Namespace{this_file.Package})
		this_file.Services = get_service_data(desc_file_info.Service, this_file)

		for _, extension := range desc_file_info.Extension {
			this_extension := new(docdata.FileExtension)
			this_extension.Name = extension.GetName()

			this_extension.FieldNumber = extension.GetNumber()

			if extension.Type != nil {
				this_extension.Type =
					field_type_enum_to_string(*extension.Type)
			}

			this_extension.Extendee = extension.GetExtendee()

			this_file.Extensions =
				append(this_file.Extensions, this_extension)
		}

		if desc_file_info.SourceCodeInfo != nil {
			add_source_code_info(this_file, desc_file_info.SourceCodeInfo)
		}
	}

	extensions.ProcessExtensions(template_data, file_descriptors, plugin_opts)

	return template_data, nil
}

func get_service_data(
	desc_services []*desc_pb.ServiceDescriptorProto,
	file_data *docdata.FileData,
) []*docdata.ServiceData {
	svc_data := make([]*docdata.ServiceData, 0, len(desc_services))
	for _, desc := range desc_services {
		this_svc := new(docdata.ServiceData)
		this_svc.Name = desc.GetName()
		this_svc.FullName = file_data.Package + "." + this_svc.Name

		methods := make([]*docdata.MethodData, 0, len(desc.Method))
		for _, method := range desc.Method {
			methods = append(methods,
				get_method_data_from_desc(method, this_svc, file_data))
		}
		this_svc.Methods = methods

		this_svc.Options = desc.Options

		svc_data = append(svc_data, this_svc)
	}

	return svc_data
}

func get_method_data_from_desc(
	desc_method *desc_pb.MethodDescriptorProto,
	svc_data *docdata.ServiceData,
	file_data *docdata.FileData,
) *docdata.MethodData {
	method_data := new(docdata.MethodData)
	method_data.Name = desc_method.GetName()
	method_data.FullName = svc_data.FullName + "." + method_data.Name
	method_data.RequestType, method_data.RequestFullType =
		extract_type_names(desc_method.GetInputType())
	method_data.RequestStreaming = desc_method.GetClientStreaming()
	method_data.ResponseType, method_data.ResponseFullType =
		extract_type_names(desc_method.GetOutputType())
	method_data.ResponseStreaming = desc_method.GetServerStreaming()
	method_data.Options = desc_method.GetOptions()

	return method_data
}

func extract_type_names(
	element_type string,
) (unqualified_type string, the_full_type string) {
	the_full_type = strings.TrimPrefix(element_type, ".")
	parts := strings.Split(element_type, ".")
	unqualified_type = parts[len(parts)-1]

	return
}

func add_source_code_info(
	file_data *docdata.FileData,
	source_info *desc_pb.SourceCodeInfo,
) {
	for _, location := range source_info.Location {
		loc_path := location.Path
		if loc_path == nil {
			continue
		}
		desc_field_num := loc_path[0]

		switch desc_field_num {
		case 4: // message
			msg := file_data.Messages[loc_path[1]]
			add_msg_desc(loc_path[2:], msg, location)
		case 5: // enum
			this_enum := file_data.Enums[loc_path[1]]
			add_enum_comments(loc_path[2:], this_enum, location)
		case 6: // service
			svc := file_data.Services[loc_path[1]]
			add_service_comments(loc_path, svc, location)
		case 12: // syntax
			syntax := file_data.Syntax
			syntax.LeadingDetachedComments = location.LeadingDetachedComments
			syntax.LeadingComments, syntax.TrailingComments,
				syntax.Description = clean_comments(location)
		}

	}
}

func clean_comments(
	location *desc_pb.SourceCodeInfo_Location,
) (leading_comments, trailing_comments, description string) {
	leading_comments = strings.TrimSpace(location.GetLeadingComments())
	trailing_comments = strings.TrimSpace(location.GetTrailingComments())
	description = get_description(leading_comments, trailing_comments)

	return
}

func get_description(leading_comments, trailing_comments string) string {
	desc := ""
	if leading_comments != "" {
		desc = leading_comments
	}

	if trailing_comments != "" {
		if desc != "" {
			desc += " "
		}
		desc += trailing_comments
	}

	return desc
}

func add_enum_comments(
	loc_path []int32,
	enum_data *docdata.EnumData,
	location *desc_pb.SourceCodeInfo_Location,
) {
	if len(loc_path) == 0 {
		// Comments for the enum declaration itself.
		enum_data.LeadingDetachedComments = location.LeadingDetachedComments
		enum_data.LeadingComments, enum_data.TrailingComments,
			enum_data.Description = clean_comments(location)
		return
	}

	if loc_path[0] == 2 {
		// Comments for an enum value.
		if len(loc_path) == 2 {
			enum_val := enum_data.Values[loc_path[1]]
			enum_val.LeadingComments, enum_val.TrailingComments,
				enum_val.Description = clean_comments(location)
		}
	}

}

func add_service_comments(
	loc_path []int32,
	svc *docdata.ServiceData,
	location *desc_pb.SourceCodeInfo_Location,
) {
	if len(loc_path) == 2 {
		svc.LeadingDetachedComments = location.LeadingDetachedComments
		svc.LeadingComments, svc.TrailingComments, svc.Description =
			clean_comments(location)
		return
	}

	// Service Method.
	if loc_path[2] == 2 {
		if len(loc_path) == 4 {
			method := svc.Methods[loc_path[3]]
			method.LeadingDetachedComments = location.LeadingDetachedComments
			method.LeadingComments, method.TrailingComments,
				method.Description = clean_comments(location)
		}

	}
}

func add_msg_desc(
	loc_path []int32,
	msg *docdata.MessageData,
	location *desc_pb.SourceCodeInfo_Location,
) {
	if len(loc_path) == 0 {
		msg.LeadingComments, msg.TrailingComments, msg.Description =
			clean_comments(location)
		return
	}

	slot_num := loc_path[0]
	switch slot_num {
	case 2:
		// Field comment.
		field := msg.Fields[loc_path[1]]
		if len(loc_path) == 2 {
			field.LeadingComments, field.TrailingComments, field.Description =
				clean_comments(location)
			field.LeadingDetachedComments = location.LeadingDetachedComments
		}
	case 3:
		// Nested messages
		nested_msg := msg.NestedMessages[loc_path[1]]
		add_msg_desc(loc_path[2:], nested_msg, location)
	case 4:
		// Enum within a message.
		enum_data := msg.Enums[loc_path[1]]
		loc_path = loc_path[2:]
		add_enum_comments(loc_path, enum_data, location)
	case 8:
		// Oneof declaration.

	}
}

func get_msg_data_from_desc(
	msg *desc_pb.DescriptorProto,
	namespace docdata.Namespace,
) *docdata.MessageData {
	this_msg := new(docdata.MessageData)

	this_msg.Name = msg.GetName()
	this_msg.FullName = namespace.QualifyName(this_msg.Name)

	msg_ns := namespace.Extend(this_msg.Name)

	fields := make([]*docdata.FieldData, 0, len(msg.Field))
	for _, field_info := range msg.Field {
		fields = append(fields, get_field_data_from_desc(field_info, msg_ns))
	}
	this_msg.Fields = fields

	this_msg.Enums = get_enum_data(msg.EnumType, namespace)

	this_msg.OneofDecls = get_oneof_data(msg.OneofDecl, msg_ns)
	this_msg.ExtensionRanges = msg.ExtensionRange
	this_msg.Options = msg.Options

	if len(msg.NestedType) > 0 {
		for _, nested_msg := range msg.NestedType {
			this_msg.NestedMessages = append(this_msg.NestedMessages,
				get_msg_data_from_desc(nested_msg, msg_ns))
		}
	}

	return this_msg
}

func get_oneof_data(
	desc_oneof_decls []*desc_pb.OneofDescriptorProto,
	namespace docdata.Namespace,
) []*docdata.OneOfData {
	oneofs := make([]*docdata.OneOfData, 0, len(desc_oneof_decls))
	for _, oneof_decl := range desc_oneof_decls {
		this_oneof := &docdata.OneOfData{
			Name:     oneof_decl.GetName(),
			FullName: namespace.QualifyName(oneof_decl.GetName()),
		}

		oneofs = append(oneofs, this_oneof)
	}

	return oneofs
}

func get_enum_data(
	desc_enums []*desc_pb.EnumDescriptorProto,
	namespace docdata.Namespace,
) []*docdata.EnumData {
	enum_data := []*docdata.EnumData{}

	if len(desc_enums) == 0 {
		return enum_data
	}

	for _, desc_enum := range desc_enums {
		this_enum := new(docdata.EnumData)
		enum_data = append(enum_data, this_enum)
		this_enum.Name = desc_enum.GetName()
		this_enum.FullName = namespace.QualifyName(this_enum.Name)
		log.Debugf("found enum %q", this_enum.Name)

		for _, value := range desc_enum.Value {
			this_val := new(docdata.EnumValue)
			if value.Name != nil {
				this_val.Name = *value.Name
			}
			if value.Number != nil {
				this_val.Number = *value.Number
			}

			this_val.Options = value.Options

			this_enum.Values = append(this_enum.Values, this_val)
		}

		this_enum.Options = desc_enum.Options
	}

	return enum_data
}

func get_field_data_from_desc(
	field *desc_pb.FieldDescriptorProto,
	namespace docdata.Namespace,
) *docdata.FieldData {
	this_field := new(docdata.FieldData)

	this_field.Name = field.GetName()
	this_field.FullName = namespace.QualifyName(this_field.Name)
	this_field.FieldNumber = field.GetNumber()

	if field.Label != nil {
		s := desc_pb.FieldDescriptorProto_Label_name[int32(*field.Label)]
		s = strings.ToLower(s)
		this_field.Label = strings.TrimPrefix(s, "label_")
	}
	if field.TypeName != nil {
		// FIXME: handle scoping here to provide full name in all cases.
		this_field.TypeName = field.GetTypeName()
		this_field.TypeName, this_field.FullTypeName =
			extract_type_names(field.GetTypeName())
	}

	if field.Type != nil {
		this_field.Kind = field_type_enum_to_string(field.GetType())
		if this_field.TypeName == "" {
			this_field.TypeName = this_field.Kind
			this_field.FullTypeName = this_field.Kind
		}
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
