package utils

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"
)

func TestNewStringArgs(t *testing.T) {
	args := NewStringArgs()
	if len(args.UniqueMap) != 0 {
		t.Errorf("Expected new StringArgs to have an empty UniqueMap, got %v", args.UniqueMap)
	}
}

func TestStringArgs_Add(t *testing.T) {
	args := NewStringArgs()
	args.Add("test")

	if _, exists := args.UniqueMap["test"]; !exists {
		t.Errorf("Expected 'test' to be added to UniqueMap, but it wasn't")
	}
}

func TestStringArgs_Set(t *testing.T) {
	args := NewStringArgs()
	args.Set("key", "value")

	if val, exists := args.UniqueMap["key"]; !exists || val != "value" {
		t.Errorf("Expected 'key' to be set to 'value' in UniqueMap, but got %v", args.UniqueMap["key"])
	}
}

func TestStringArgs_Diff(t *testing.T) {
	args1 := NewStringArgs()
	args1.Add("a")
	args1.Add("b")
	args1.Add("c")

	args2 := NewStringArgs()
	args2.Add("b")
	args2.Add("c")

	expectedDiff := NewStringArgs()
	expectedDiff.Add("a")

	diff := args1.Diff(args2)
	if !reflect.DeepEqual(diff.UniqueMap, expectedDiff.UniqueMap) {
		t.Errorf("Expected diff to be %v, but got %v", expectedDiff.UniqueMap, diff.UniqueMap)
	}
}

func TestStringArgs_ToString(t *testing.T) {
	args := NewStringArgs()
	args.Add("a")
	args.Add("b")
	args.Add("c")

	expected := []string{"a", "b", "c"}
	result := args.ToString()

	if diff := deep.Equal(expected, result, deep.FLAG_IGNORE_SLICE_ORDER); diff != nil {
		t.Errorf("Expected ToString to return %v, but got %v", expected, result)
	}
}

func TestStringArgs_Args(t *testing.T) {
	args := NewStringArgs()
	args.Add("a")
	args.Add("b")
	args.Add("c")

	expected := []interface{}{"a", "b", "c"}
	result := args.Args()

	if diff := deep.Equal(expected, result, deep.FLAG_IGNORE_SLICE_ORDER); diff != nil {
		t.Errorf("Expected Args to return %v, but got %v", expected, result)
	}
}
