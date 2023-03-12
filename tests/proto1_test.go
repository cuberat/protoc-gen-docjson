package proto1_test

import (
	// Built-in/core modules.
	"encoding/json"
	"fmt"
	"io"
	"os"
	exec "os/exec"
	"path"
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

	t.Run("extension check",
		func(st *testing.T) {
			do_check_extensions(st, data)
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

	_ = svc
	// FIXME: complete service checks: comments on service and methods.

	// for _, svc_name_any := range svc_name_list {
	// 	svc_name, ok := svc_name_any.(string)
	// 	if !ok {
	// 		t.Errorf("Wrong type for service name: %T", svc_name_any)
	// 		return
	// 	}
	// 	svc, ok := svc_map[svc_name].(map[string]any)
	// 	if !ok {
	// 		t.Errorf("Wrong type for service: %T", svc_map["service_name"])
	// 		return
	// 	}

	// 	_ = svc
	// }
}

func do_check_files(t *testing.T, data map[string]any) {
	// FIXME: complete file checks, e.g., extensions.
}

func do_check_messages(t *testing.T, data map[string]any) {
	// FIXME: complete message checks.
}

func do_check_extensions(t *testing.T, data map[string]any) {
	// FIXME: complete extension checks.
}

func do_check_enums(t *testing.T, data map[string]any) {
	// FIXME: complete extension checks.
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

	check_comment_field(t, syntax_data, field_desc, "leading_comments",
		"This is the syntax statement leading comment.")
	check_comment_field(t, syntax_data, field_desc, "trailing_comments",
		"This is a trailing comment for syntax.")
	check_comment_field(t, syntax_data, field_desc, "description",
		"This is the syntax statement leading comment. This is a trailing comment for syntax.")
	check_comment_list(t, syntax_data, field_desc, "leading_detached_comments",
		[]string{"Leading detached comment for the syntax statement.\n A second line."})
}

func check_comment_list(
	t *testing.T, data map[string]any,
	field_desc, field_name string,
	expected []string,
) {
	comment_list, ok := data[field_name].([]any)
	if !ok {
		t.Errorf("Wrong type for %s %s field: %T", field_desc, field_name,
			data[field_name])
	}
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
	t *testing.T, data map[string]any,
	field_desc, field_name, expected string,
) {
	comment, ok := data[field_name].(string)
	if !ok {
		t.Errorf("Wrong type for %s %s field: %T", field_desc,
			field_name, data[field_name])
		return
	}

	if comment != expected {
		t.Errorf("comment did not match for %s %s field. Got %q, expected %q.",
			field_desc, field_name, comment, expected)
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
