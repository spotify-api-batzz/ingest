package utils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-test/deep"
)

// Convert a slice of slices of strings to a slice of slices of interfaces
func convertToInterfaceSlice(s [][]string) [][]interface{} {
	result := make([][]interface{}, len(s))
	for i, subSlice := range s {
		interfaceSlice := make([]interface{}, len(subSlice))
		for j, element := range subSlice {
			interfaceSlice[j] = element
		}
		result[i] = interfaceSlice
	}
	return result
}

func TestChunkSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		size     int
		expected [][]string
	}{
		{
			name:     "Chunk evenly divisible",
			input:    []string{"a", "b", "c", "d"},
			size:     2,
			expected: [][]string{{"a", "b"}, {"c", "d"}},
		},
		{
			name:     "Chunk not evenly divisible",
			input:    []string{"a", "b", "c", "d", "e"},
			size:     2,
			expected: [][]string{{"a", "b"}, {"c", "d"}, {"e"}},
		},
		{
			name:     "Chunk size larger than slice",
			input:    []string{"a", "b"},
			size:     5,
			expected: [][]string{{"a", "b"}},
		},
		{
			name:     "Chunk size of 1",
			input:    []string{"a", "b", "c"},
			size:     1,
			expected: [][]string{{"a"}, {"b"}, {"c"}},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			size:     3,
			expected: [][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ChunkSlice(tt.input, tt.size)

			// Convert expected and result slices to slices of interfaces
			expectedInterfaces := convertToInterfaceSlice(tt.expected)
			resultInterfaces := convertToInterfaceSlice(result)

			// Compare slices of interfaces ignoring slice order
			if diff := deep.Equal(resultInterfaces, expectedInterfaces); diff != nil {
				fmt.Printf("ChunkSlice(%v, %d) = %v, want %v\n", tt.input, tt.size, result, tt.expected)
				t.Errorf("ChunkSlice diff %v", diff)
			}
		})
	}
}

func TestRemoveExcludedFromSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		exclude  []string
		expected []string
	}{
		{
			name:     "Test with exclusion of multiple elements",
			slice:    []string{"apple", "banana", "orange", "grape"},
			exclude:  []string{"banana", "orange"},
			expected: []string{"apple", "grape"},
		},
		{
			name:     "Test with exclusion of one element",
			slice:    []string{"apple", "banana", "orange", "grape"},
			exclude:  []string{"banana"},
			expected: []string{"apple", "orange", "grape"},
		},
		{
			name:     "Test with empty exclusion list",
			slice:    []string{"apple", "banana", "orange", "grape"},
			exclude:  []string{},
			expected: []string{"apple", "banana", "orange", "grape"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := RemoveExcludedFromSlice(test.slice, test.exclude)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}
func TestKeepIncludedInSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		include  []string
		expected []string
	}{
		{
			name:     "Test with inclusion of multiple elements",
			slice:    []string{"apple", "banana", "orange", "grape"},
			include:  []string{"banana", "orange"},
			expected: []string{"banana", "orange"},
		},
		{
			name:     "Test with inclusion of one element",
			slice:    []string{"apple", "banana", "orange", "grape"},
			include:  []string{"banana"},
			expected: []string{"banana"},
		},
		{
			name:     "Test with empty inclusion list",
			slice:    []string{"apple", "banana", "orange", "grape"},
			include:  []string{},
			expected: []string{},
		},
		{
			name:     "Test with inclusion of all elements",
			slice:    []string{"apple", "banana", "orange", "grape"},
			include:  []string{"apple", "banana", "orange", "grape"},
			expected: []string{"apple", "banana", "orange", "grape"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := KeepIncludedInSlice(test.slice, test.include)
			if diff := deep.Equal(result, test.expected); diff != nil {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}
