package proto1_test

import (
	// Built-in/core modules.
	"reflect"
	"testing"

	// First-party modules.
	util "github.com/cuberat/protoc-gen-docjson/internal/util"
)

func TestStringSet(t *testing.T) {
	expected := []string{"foo", "bar", "beefc0de"}
	set := util.NewStringSet()
	set.Add("foo")
	set.Add("bar")
	set.Add("foo")
	set.Add("beefc0de")
	set.Add("bar")
	set.Add("beefc0de")

	got := set.GetItems()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("StringSet items incorrect: got %v, expected %v",
			got, expected)
		return
	}
}

func TestIntSet(t *testing.T) {
	expected := []int{2, 1, 3}
	set := util.NewIntSet()
	set.Add(2)
	set.Add(1)
	set.Add(3)
	set.Add(2)
	set.Add(3)
	set.Add(1)

	got := set.GetItems()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("IntSet items incorrect: got %v, expected %v",
			got, expected)
		return
	}
}
