package proto1_test

import (
	// Built-in/core modules.
	"encoding/json"
	"fmt"
	"io"
	"os"
	exec "os/exec"
	"path"
	"reflect"
	"sort"
	"testing"
	// Generated code.
	// First-party modules.
)

func TestComments(t *testing.T) {
	data, ok := do_setup(t)
	if !ok {
		return
	}

	if !t.Run("len check",
		func(st *testing.T) {
			do_len_checks(st, data)
		},
	) {
		t.Error("Len checks failed. Bailing out of tests.")
		return
	}

	t.Run("syntax data check",
		func(st *testing.T) {
			do_check_syntax(st, data)
		},
	)

	t.Run("svc check",
		func(st *testing.T) {
			do_check_services(st, data)
		},
	)

	t.Run("file check",
		func(st *testing.T) {
			do_check_files(st, data)
		},
	)

	t.Run("msg check",
		func(st *testing.T) {
			do_check_messages(st, data)
		},
	)

	t.Run("enum check",
		func(st *testing.T) {
			do_check_enums(st, data)
		},
	)
}

func do_check_services(t *testing.T, data map[string]any) {
	svc_map, ok := data["service_map"].(map[string]any)
	if !ok {
		t.Errorf("Wrong type for service_map: %T", data["service_map"])
		return
	}

	svc_name := "MyServices.Service.Tester"
	svc, ok := svc_map[svc_name].(map[string]any)
	if !ok {
		t.Errorf("Wrong type for service: %T", svc_map["service_name"])
		return
	}

	label := fmt.Sprintf("service %s", svc["name"].(string))

	exp_desc := "Tester service. Lorem ipsum dolor sit amet, consectetur adipiscing elit.\n Suspendisse a cursus mauris. Proin porta mi nisl, vel iaculis leo mattis\n ut. Maecenas lacus urna, dapibus sit amet leo id, rutrum fermentum justo.\n Cras porta, nulla vel euismod maximus, lacus magna ultrices metus, sit amet\n eleifend lacus libero et lacus. Cras a facilisis est. Praesent augue nisl,\n tincidunt vel ex mattis, efficitur fermentum sem. Ut congue tellus ut\n accumsan condimentum. Sed quis leo nec turpis maximus molestie quis sit\n amet erat."

	desc := svc["description"].(string)
	if desc != exp_desc {
		t.Errorf("incorrect description for service %s. Got %q, expected %q",
			svc_name, desc, exp_desc)
		return
	}

	leading_comments := svc["leading_comments"].(string)
	if leading_comments != exp_desc {
		t.Errorf("incorrect leading_comments for service %s. "+
			"Got %q, expected %q", svc_name, leading_comments, exp_desc)
		return
	}

	leading_detached_comments := svc["leading_detached_comments"].([]any)
	if len(leading_detached_comments) != 0 {
		t.Errorf("got %d leading_detached_comments when 0 were expected",
			len(leading_detached_comments))
		return
	}

	check_methods(t, svc)

	test_spec := map[string]any{
		"options": map[string]any{
			"deprecated": false,
		},
		"custom_options": map[string]any{
			"service_not_implemented": true,
		},
	}

	t.Run("options",
		func(st *testing.T) {
			do_check_options(st, svc, test_spec, label)
		},
	)
}

func check_one_method(
	t *testing.T,
	this_method map[string]any,
	test_spec map[string]any,
	label string,
) {
	method_name := this_method["name"].(string)
	method_full_name := this_method["full_name"].(string)

	exp_name := test_spec["name"].(string)
	if method_name != exp_name {
		t.Errorf("got method name %q, expected %q", method_name, exp_name)
	}
	exp_full_name := test_spec["full_name"]
	if method_full_name != exp_full_name {
		t.Errorf("got method full name %q, expected %q", method_full_name,
			exp_full_name)
	}

	desc := this_method["description"].(string)
	exp_desc := test_spec["desc"]

	if desc != exp_desc {
		t.Errorf("incorrect description for service method %s: got %q, "+
			"expected %q", method_name, desc, exp_desc)
	}

	t.Run("request_response",
		func(st *testing.T) {
			check_req_resp(st, this_method, test_spec, label)
		},
	)

	t.Run("options",
		func(st *testing.T) {
			do_check_options(st, this_method, test_spec, label)
		},
	)

	defined_in := test_spec["defined_in"].(string)
	if file_name := this_method["defined_in"].(string); file_name != defined_in {
		t.Errorf("incorrect file name in `defined_in` field: got %q, "+
			"expected %q", this_method["defined_in"].(string), defined_in)
	}

}

func check_req_resp(
	t *testing.T,
	this_method map[string]any,
	test_spec map[string]any,
	label string,
) {
	exp_request_type := test_spec["request_type"].(string)
	got_request_type := this_method["request_type"].(string)
	if got_request_type != exp_request_type {
		t.Errorf("incorrect request type for %s: got %q, expected %q", label,
			got_request_type, exp_request_type)
	}
	exp_request_type = test_spec["request_full_type"].(string)
	got_request_type = this_method["request_full_type"].(string)
	if got_request_type != exp_request_type {
		t.Errorf("incorrect request full type for %s: got %q, expected %q",
			label, got_request_type, exp_request_type)
	}

	exp_response_type := test_spec["response_type"].(string)
	got_response_type := this_method["response_type"].(string)
	if got_response_type != exp_response_type {
		t.Errorf("incorrect response type for %s: got %q, expected %q",
			label, got_response_type, exp_response_type)
	}

	exp_response_type = test_spec["response_full_type"].(string)
	got_response_type = this_method["response_full_type"].(string)
	if got_response_type != exp_response_type {
		t.Errorf("incorrect response full type for %s: got %q, expected %q",
			label, got_response_type, exp_response_type)
	}
}

// Performs unit tests for service methods. Returns true if all tests pass.
// Returns false, otherwise.
func check_methods(t *testing.T, svc map[string]any) {
	methods := svc["methods"].([]any)
	if len(methods) != 2 {
		t.Fatalf("got %d methods, expected 2", len(methods))
	}

	method_test_spec := []map[string]any{
		{
			"name":       "RunTestV2",
			"full_name":  "MyServices.Service.Tester.RunTestV2",
			"desc":       "Leading comment for the RunTestV2 method which is marked not_implemented\n via a custom option method_not_implemented.",
			"defined_in": "service-tester.proto",
			"options": map[string]any{
				"deprecated": false,
			},
			"custom_options": map[string]any{
				"method_not_implemented": true,
			},
			"request_type":       "TesterRequest",
			"request_full_type":  "MyServices.Tester.TesterRequest",
			"response_type":      "TesterResponse",
			"response_full_type": "MyServices.Tester.TesterResponse",
		},

		{
			"name":       "RunTest",
			"full_name":  "MyServices.Service.Tester.RunTest",
			"desc":       "Leading comment for the RunTest method is which marked deprecated.",
			"defined_in": "service-tester.proto",
			"options": map[string]any{
				"deprecated": true,
			},
			"custom_options":     map[string]any{},
			"request_type":       "TesterRequest",
			"request_full_type":  "MyServices.Tester.TesterRequest",
			"response_type":      "TesterResponse",
			"response_full_type": "MyServices.Tester.TesterResponse",
		},
	}

	for i, this_method_any := range methods {
		this_method := this_method_any.(map[string]any)
		this_test_spec := method_test_spec[i]
		this_method_label := fmt.Sprintf("method %s",
			this_method["name"].(string))

		t.Run(this_method_label, func(st *testing.T) {
			check_one_method(st, this_method, this_test_spec, this_method_label)
		})
	}
}

func do_check_options(
	t *testing.T,
	data map[string]any,
	test_spec map[string]any,
	label string,
) {
	// Check standard options.
	std_options := data["options"].(map[string]any)
	std_options_test := test_spec["options"].(map[string]any)
	for _, opt_name := range get_sorted_keys(std_options_test) {
		opt_val_exp := std_options_test[opt_name]
		opt_val_got := std_options[opt_name]
		if !reflect.DeepEqual(opt_val_exp, opt_val_got) {
			t.Logf("incorrect value for standard option %q in %s: "+
				"got %v, expected %v", label, opt_name, opt_val_got,
				opt_val_exp)
		}
	}

	// Check custom options.
	cust_options := data["custom_options"].(map[string]any)
	cust_options_test := test_spec["custom_options"].(map[string]any)
	for _, opt_name := range get_sorted_keys(cust_options_test) {
		opt_val_exp := cust_options_test[opt_name]
		opt_val_got := cust_options[opt_name]
		if !reflect.DeepEqual(opt_val_exp, opt_val_got) {
			t.Logf("incorrect value for custom option %q in %s: "+
				"got %v, expected %v", label, opt_name, opt_val_got,
				opt_val_exp)
		}
	}
}

func check_one_file(
	t *testing.T,
	this_file, test_spec map[string]any,
	label string,
) {
	check_extensions(t, this_file, test_spec, label)
	do_check_options(t, this_file, test_spec, label)
	field_check_list := []string{
		"version", "name", "package", "defined_in",
	}
	check_fields_equal(t, this_file, test_spec, label, field_check_list)
}

func do_check_files(t *testing.T, data map[string]any) {
	file_test_spec := map[string]any{
		"service-tester.proto": map[string]any{
			"extensions": []map[string]any{},
			"name":       "service-tester.proto",
			"package":    "MyServices.Service",
			"messages":   []map[string]any{},
			"enums":      []map[string]any{},
			"options": map[string]any{
				"java_package":           "myservices.service",
				"java_outer_classname":   "",
				"java_multiple_files":    false,
				"java_string_check_utf8": false,
				"go_package":             "myservices/grpc/service",
				"deprecated":             false,
				"cc_enable_arenas":       true,
				"objc_class_prefix":      "",
				"csharp_namespace":       "",
				"swift_prefix":           "",
				"php_class_prefix":       "",
				"php_namespace":          "",
				"php_metadata_namespace": "",
				"ruby_package":           "",
			},
			"custom_options": map[string]any{},
		},
		"tester.proto": map[string]any{
			"description":               "Tester main structure, TesterRequest.",
			"leading_comments":          "Tester main structure, TesterRequest.",
			"trailing_comments":         "",
			"leading_detached_comments": map[string]any{},
			"name":                      "tester.proto",
			"package":                   "MyServices.Tester",
			"options": map[string]any{
				"java_package":           "Myservices.tester",
				"java_outer_classname":   "",
				"java_multiple_files":    false,
				"java_string_check_utf8": false,
				"go_package":             "Myservices/grpc/tester",
				"deprecated":             false,
				"cc_enable_arenas":       true,
				"objc_class_prefix":      "",
				"csharp_namespace":       "",
				"swift_prefix":           "",
				"php_class_prefix":       "",
				"php_namespace":          "",
				"php_metadata_namespace": "",
				"ruby_package":           "",
			},
			"custom_options": map[string]any{
				"file_deprecated": true,
				"file_double":     float64(5643343.423),
				"file_float":      float64(0.569),
				"file_int64":      float64(-343434343),
				"file_mnemonic":   "some random name",
			},
			"extensions": []map[string]any{
				{
					"description":               "Leading comment for custom option field_required.",
					"leading_comments":          "Leading comment for custom option field_required.",
					"trailing_comments":         "",
					"leading_detached_comments": []any{},
					"name":                      "field_required",
					"full_name":                 "MyServices.Tester.field_required",
					"field_number":              float64(51234),
					"type":                      "bool",
					"extendee":                  ".google.protobuf.FieldOptions",
					"defined_in":                "tester.proto",
				},
				{
					"description":               "Trailing for method_not_implemented.",
					"leading_comments":          "",
					"trailing_comments":         "Trailing for method_not_implemented.",
					"leading_detached_comments": []any{},
					"name":                      "method_not_implemented",
					"full_name":                 "MyServices.Tester.method_not_implemented",
					"field_number":              float64(51235),
					"type":                      "bool",
					"extendee":                  ".google.protobuf.MethodOptions",
					"defined_in":                "tester.proto",
				},
				{
					"description":               "",
					"leading_comments":          "",
					"trailing_comments":         "",
					"leading_detached_comments": []any{},
					"name":                      "message_not_implemented",
					"full_name":                 "MyServices.Tester.message_not_implemented",
					"field_number":              float64(51236),
					"type":                      "bool",
					"extendee":                  ".google.protobuf.MessageOptions",
					"defined_in":                "tester.proto",
				},
				{
					"description":               "",
					"leading_comments":          "",
					"trailing_comments":         "",
					"leading_detached_comments": []any{},
					"name":                      "service_not_implemented",
					"full_name":                 "MyServices.Tester.service_not_implemented",
					"field_number":              float64(51237),
					"type":                      "bool",
					"extendee":                  ".google.protobuf.ServiceOptions",
					"defined_in":                "tester.proto",
				},
				{
					"description":               "",
					"leading_comments":          "",
					"trailing_comments":         "",
					"leading_detached_comments": []any{},
					"name":                      "file_deprecated",
					"full_name":                 "MyServices.Tester.file_deprecated",
					"field_number":              float64(51238),
					"type":                      "bool",
					"extendee":                  ".google.protobuf.FileOptions",
					"defined_in":                "tester.proto",
				},
				{
					"description":               "",
					"leading_comments":          "",
					"trailing_comments":         "",
					"leading_detached_comments": []any{},
					"name":                      "file_mnemonic",
					"full_name":                 "MyServices.Tester.file_mnemonic",
					"field_number":              float64(51240),
					"type":                      "string",
					"extendee":                  ".google.protobuf.FileOptions",
					"defined_in":                "tester.proto",
				},
				{
					"description":               "",
					"leading_comments":          "",
					"trailing_comments":         "",
					"leading_detached_comments": []any{},
					"name":                      "file_double",
					"full_name":                 "MyServices.Tester.file_double",
					"field_number":              float64(51241),
					"type":                      "float",
					"extendee":                  ".google.protobuf.FileOptions",
					"defined_in":                "tester.proto",
				},
				{
					"description":               "",
					"leading_comments":          "",
					"trailing_comments":         "",
					"leading_detached_comments": []any{},
					"name":                      "file_float",
					"full_name":                 "MyServices.Tester.file_float",
					"field_number":              float64(51242),
					"type":                      "float",
					"extendee":                  ".google.protobuf.FileOptions",
					"defined_in":                "tester.proto",
				},
				{
					"description":               "",
					"leading_comments":          "",
					"trailing_comments":         "",
					"leading_detached_comments": []any{},
					"name":                      "file_int64",
					"full_name":                 "MyServices.Tester.file_int64",
					"field_number":              float64(51243),
					"type":                      "float",
					"extendee":                  ".google.protobuf.FileOptions",
					"defined_in":                "tester.proto",
				},
				{
					"description":               "",
					"leading_comments":          "",
					"trailing_comments":         "",
					"leading_detached_comments": []any{},
					"name":                      "enum_deprecated",
					"full_name":                 "MyServices.Tester.enum_deprecated",
					"field_number":              float64(51239),
					"type":                      "bool",
					"extendee":                  ".google.protobuf.EnumOptions",
					"defined_in":                "tester.proto",
				},
			},
		},
		"subdir/docstuff.proto": map[string]any{
			"extensions": []map[string]any{},
			"name":       "subdir/docstuff.proto",
			"package":    "MyServices.Tester.Subdir",
			"options": map[string]any{
				"java_package":           "",
				"java_outer_classname":   "",
				"java_multiple_files":    false,
				"java_string_check_utf8": false,
				"go_package":             "",
				"deprecated":             false,
				"cc_enable_arenas":       true,
				"objc_class_prefix":      "",
				"csharp_namespace":       "",
				"swift_prefix":           "",
				"php_class_prefix":       "",
				"php_namespace":          "",
				"php_metadata_namespace": "",
				"ruby_package":           "",
			},
			"custom_options":          map[string]any{},
			"declared_custom_options": map[string]any{},
		},
	}

	file_map := data["file_map"].(map[string]any)
	file_names := get_sorted_keys(file_map)
	t.Logf("file_names: %v", file_names)
	for _, file_name := range file_names {
		t.Logf("file_name: %s", file_name)
		file_data := file_map[file_name].(map[string]any)
		label := fmt.Sprintf("file %s", file_name)
		t.Run(label, func(st *testing.T) {
			check_one_file(st, file_data,
				file_test_spec[file_name].(map[string]any), label)
		})
	}
}

func do_check_messages(t *testing.T, data map[string]any) {
	msg_map := data["message_map"].(map[string]any)
	msg_full_name := "MyServices.Tester.TesterRequest"
	msg := msg_map[msg_full_name].(map[string]any)
	fields := msg["fields"].([]any)

	exp_msg_comments := map[string]any{
		"description":               "Tester main structure, TesterRequest.",
		"leading_comments":          "Tester main structure, TesterRequest.",
		"trailing_comments":         "",
		"leading_detached_comments": []string{},
	}

	check_comments(t, msg, exp_msg_comments,
		fmt.Sprintf("msg %s", msg_full_name))

	field := fields[0].(map[string]any)

	options := field["options"]
	custom_options := field["custom_options"]
	exp_field_comments := map[string]any{
		"description":               "Leading comment for the client_info field.",
		"leading_comments":          "Leading comment for the client_info field.",
		"trailing_comments":         "",
		"leading_detached_comments": []string{},
	}

	check_comments(t, field, exp_field_comments,
		fmt.Sprintf("field number %d of %s",
			field["field_number"], msg_full_name))

	exp_options := map[string]any{
		"ctype":      0,
		"packed":     false,
		"jstype":     0,
		"lazy":       false,
		"deprecated": false,
	}
	exp_custom_options := map[string]any{
		"field_required": true,
	}

	if !reflect.DeepEqual(options, exp_options) {
		t.Logf("options in first field of %s incorrect: "+
			"got %v, expected %v", msg_full_name, options, exp_options)
	}

	if !reflect.DeepEqual(custom_options, exp_custom_options) {
		t.Logf("custom options in first field of %s incorrect: "+
			"got %v, expected %v", msg_full_name, custom_options,
			exp_custom_options)
	}
}

func check_extensions(
	t *testing.T, data map[string]any,
	test_spec map[string]any,
	label string,
) {
	if data["extensions"] == nil {
		return
	}
	extension_list_any := data["extensions"].([]any)
	ext_spec_list := test_spec["extensions"].([]map[string]any)

	extension_map := make(map[string]map[string]any, len(extension_list_any))
	for _, ext_any := range extension_list_any {
		ext := ext_any.(map[string]any)
		name := ext["full_name"].(string)
		extension_map[name] = ext
	}

	if len(extension_map) != len(ext_spec_list) {
		t.Fatalf("mismatch in number of found extensions vs expected in %s: "+
			"found %d, expected %d", label, len(extension_map),
			len(ext_spec_list))
	}

	for _, ext_spec := range ext_spec_list {
		ext_name := ext_spec["full_name"].(string)
		this_label := fmt.Sprintf("extension %s in %s", ext_name, label)
		ext := extension_map[ext_name]
		if ext == nil {
			t.Fatalf("missing %s", this_label)
		}

		t.Run(this_label, func(st *testing.T) {
			check_one_extension(st, ext, ext_spec, this_label)
		})

	}
}

func check_one_extension(
	t *testing.T,
	ext, ext_test_spec map[string]any,
	label string,
) {
	check_fields_equal(t, ext, ext_test_spec, label, nil)
}

func check_fields_equal(
	t *testing.T,
	data map[string]any,
	test_spec map[string]any,
	label string,
	field_list []string) {

	if len(field_list) == 0 {
		field_list = get_sorted_keys(test_spec)
	}

	for _, field := range field_list {
		if !reflect.DeepEqual(data[field], test_spec[field]) {
			t.Errorf("discrepancy with field %s: got %v (%T), expected %v (%T)",
				field, data[field], data[field], test_spec[field], test_spec[field])
		}
	}
}

func do_check_enums(t *testing.T, data map[string]any) {
	expected := map[string]any{
		"leading_comments":  "Leading comment for enum TesterError.",
		"trailing_comments": "",
		"leading_detached_comments": []string{
			"Detached leading comment for enum TesterError",
		},
		"name":        "TesterError",
		"full_name":   "MyServices.Tester.TesterError",
		"description": "Leading comment for enum TesterError.",
		"values": []map[string]any{
			{
				"description":               "Leading comment for enum value NONE.",
				"leading_comments":          "Leading comment for enum value NONE.",
				"trailing_comments":         "",
				"leading_detached_comments": []string{},
				"Name":                      "NONE",
				"Number":                    0,
				"options": map[string]any{
					"deprecated": false,
				},
				"custom_options": map[string]any{},
			},
		},
		"options": map[string]any{
			"allow_alias": false,
			"deprecated":  false,
		},
		"custom_options": map[string]any{
			"enum_deprecated": true,
		},
		"defined_in": "tester.proto",
	}

	enum_map := data["enum_map"].(map[string]any)
	this_enum := enum_map["MyServices.Tester.TesterError"].(map[string]any)

	check_comments(t, this_enum, expected, "enum MyServices.Tester.TesterError")

	enum_vals := this_enum["values"].([]any)
	if len(enum_vals) != 3 {
		t.Errorf("incorrect number of values for enum "+
			"MyServices.Tester.TesterError: got %d, expected 3",
			len(enum_vals))
	}

	enum_val_0 := enum_vals[0].(map[string]any)
	expected_enum_vals := expected["values"].([]map[string]any)
	exp_val_0 := expected_enum_vals[0]

	check_comments(t, enum_val_0, exp_val_0,
		"val 0 from enum MyServices.Tester.TesterError")

	do_check_options(t, this_enum, expected,
		"options for enum MyServices.Tester.TesterError")

}

func do_check_syntax(t *testing.T, data map[string]any) {
	file_map, ok := data["file_map"].(map[string]any)
	if !ok {
		t.Errorf("Wrong type for file_map: %T", file_map)
		return
	}

	file_name := "tester.proto"
	tester_file, ok := file_map[file_name].(map[string]any)
	if !ok {
		t.Errorf("Wrong type for %q info: %T", file_name, file_map[file_name])
		return
	}

	syntax_data, ok := tester_file["syntax"].(map[string]any)
	if !ok {
		t.Errorf("Wrong type for syntax_data: %T", tester_file["syntax"])
		return
	}

	field_desc := fmt.Sprintf("syntax (in %q)", file_name)

	test_spec_map := map[string]any{
		"leading_comments":          "This is the syntax statement leading comment.",
		"trailing_comments":         "This is a trailing comment for syntax.",
		"description":               "This is the syntax statement leading comment. This is a trailing comment for syntax.",
		"leading_detached_comments": []string{"Leading detached comment for the syntax statement.\n A second line."},
	}

	check_comments(t, syntax_data, test_spec_map, field_desc)
}

func check_comments(
	t *testing.T,
	data, test_spec map[string]any,
	label string,
) {
	check_comment_field(t, data, label, "leading_comments",
		test_spec)
	check_comment_field(t, data, label, "trailing_comments",
		test_spec)
	check_comment_field(t, data, label, "description",
		test_spec)
	check_comment_list(t, data, label, "leading_detached_comments",
		test_spec)
}

func check_comment_list(
	t *testing.T, data map[string]any,
	field_desc, field_name string,
	test_spec map[string]any,
) {
	comment_list, ok := data[field_name].([]any)
	if !ok {
		t.Errorf("Wrong type for %s %s field: %T", field_desc, field_name,
			data[field_name])
	}
	expected := test_spec[field_name].([]string)
	if len(expected) != len(comment_list) {
		t.Errorf("Wrong length for %s %s list: got %d, expected %d",
			field_desc, field_name, len(comment_list), len(expected))
		return
	}

	for i, exp_val := range expected {
		comment, ok := comment_list[i].(string)
		if !ok {
			t.Errorf("Wrong type for comment list index %d for %s %s: %T",
				i, field_desc, field_name, comment_list[i])
			return
		}

		if comment != exp_val {
			t.Errorf("Incorrect comment at index %d for %s %s: got %q, "+
				"expected %q", i, field_desc, field_name, comment, exp_val)
		}
	}
}

func check_comment_field(
	t *testing.T,
	data map[string]any,
	label, field_name string,
	test_spec map[string]any,
) {
	comment, ok := data[field_name].(string)
	if !ok {
		t.Errorf("Wrong type for %s %s field: %T", label,
			field_name, data[field_name])
		return
	}

	expected := test_spec[field_name].(string)
	if comment != expected {
		t.Errorf("comment did not match for %s %s field. Got %q, expected %q.",
			label, field_name, comment, expected)
		return
	}
}

func do_len_checks(t *testing.T, data map[string]any) {
	if !check_length(t, data, "file_name_list", "file_map", 3) {
		return
	}
	if !check_length(t, data, "message_name_list", "message_map", 7) {
		return
	}
	if !check_length(t, data, "service_name_list", "service_map", 1) {
		return
	}
	if !check_length(t, data, "enum_name_list", "enum_map", 2) {
		return
	}
	if !check_length(t, data, "extension_name_list", "extension_map", 10) {
		return
	}

}

func get_sorted_keys(data map[string]any) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func check_length(
	t *testing.T,
	data map[string]any,
	list_field_name, map_field_name string,
	expected_len int,
) bool {
	name_list, ok := data[list_field_name].([]any)
	if !ok {
		t.Errorf("wrong type for %s: %T", list_field_name,
			data[list_field_name])
		return false
	}
	if len(name_list) != expected_len {
		t.Errorf("wrong length for %s: got %d, expected %d: list=%v",
			list_field_name, len(name_list), expected_len, name_list)
		return false
	}

	name_map, ok := data[map_field_name].(map[string]any)
	if !ok {
		t.Errorf("wrong type for %s: %T", map_field_name, data[map_field_name])
		return false
	}
	if len(name_map) != len(name_list) {
		t.Errorf("length of %s does not match that of %s: %T",
			list_field_name, map_field_name, len(name_map))
		return false
	}

	return true
}

func do_setup(t *testing.T) (map[string]any, bool) {
	cur_dir, err := os.Getwd()
	if err != nil {
		t.Errorf("couldn't get working directory: %s", err)
		return nil, false
	}
	defer os.Chdir(cur_dir)

	work_dir := path.Join(cur_dir, "data/proto1")
	proto_dir := work_dir
	bin_dir := path.Join(cur_dir, "../cmd/protoc-gen-docjson")
	out_file_name := "docs.json"

	// Could also use t.TempDir() here, which would get cleaned up automatically
	// once the test is complete.
	json_out_dir, err := os.MkdirTemp("", "test_protoc-gen-docjson_*")
	if err != nil {
		t.Errorf("couldn't create temporary directory: %s", err)
		return nil, false
	}
	defer os.RemoveAll(json_out_dir)

	json_out_path := path.Join(json_out_dir, out_file_name)

	cmd := "/usr/bin/env"
	args := []string{
		fmt.Sprintf("PATH=%s:%s", bin_dir, os.Getenv("PATH")),
		"protoc",
		fmt.Sprintf("--docjson_out=%s", json_out_dir),
		fmt.Sprintf("--docjson_opt=outfile=%s,proto=%s",
			out_file_name, proto_dir),
		fmt.Sprintf("-I%s", proto_dir),
		"tester.proto", "service-tester.proto", "subdir/docstuff.proto",
	}

	os.Chdir(work_dir)
	t.Logf("running cmd %s %s", cmd, args)
	if err := exec.Command(cmd, args...).Run(); err != nil {
		t.Errorf("protobuf compiler failed: %s", err)
		return nil, false
	}

	in_fh, err := os.Open(json_out_path)
	if err != nil {
		t.Errorf("couldn't open JSON file %q: %s", json_out_path, err)
		return nil, false
	}
	defer in_fh.Close()

	json_bytes, err := io.ReadAll(in_fh)
	if err != nil {
		t.Errorf("couldn't read JSON data from %q: %s", json_out_path, err)
		return nil, false
	}

	// Checkout JSON against expected output.
	data := make(map[string]any)
	err = json.Unmarshal(json_bytes, &data)
	if err != nil {
		t.Errorf("couldn't unmarshal JSON file into data structure: %s", err)
		return nil, false
	}

	return data, true
}
