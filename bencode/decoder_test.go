package bencode

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	// Group test cases by type for better organization
	t.Run("String tests", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:     "Valid bencoded string",
				input:    []byte("5:hello"),
				expected: "hello",
				bytes:    7,
				err:      nil,
			},
			{
				name:     "Invalid bencoded string (no colon)",
				input:    []byte("5hello"),
				expected: "",
				bytes:    0,
				err:      errors.New("invalid string format: no colon found"),
			},
			{
				name:     "String data too short",
				input:    []byte("5:hel"),
				expected: "",
				bytes:    0,
				err:      errors.New("string data too short"),
			},
		})
	})

	t.Run("Integer tests", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:     "Valid bencoded integer",
				input:    []byte("i42e"),
				expected: int64(42),
				bytes:    4,
				err:      nil,
			},
			{
				name:     "Invalid bencoded integer (no end marker)",
				input:    []byte("i42"),
				expected: int64(0),
				bytes:    0,
				err:      errors.New("invalid integer format: no end marker"),
			},
			{
				name:     "Invalid bencoded integer (leading zeros)",
				input:    []byte("i042e"),
				expected: int64(0),
				bytes:    0,
				err:      errors.New("invalid integer format: leading zeros"),
			},
			{
				name:     "Invalid bencoded integer (negative zero)",
				input:    []byte("i-0e"),
				expected: int64(0),
				bytes:    0,
				err:      errors.New("invalid integer format: negative zero"),
			},
			{
				name:     "Invalid bencoded integer (non-numeric)",
				input:    []byte("i4a2e"),
				expected: int64(0),
				bytes:    0,
				err:      fmt.Errorf("invalid integer: strconv.ParseInt: parsing \"4a2\": invalid syntax"),
			},
		})
	})

	t.Run("List tests", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:     "Valid bencoded list",
				input:    []byte("li1ei2ei3ee"),
				expected: []interface{}{int64(1), int64(2), int64(3)},
				bytes:    11,
				err:      nil,
			},
			{
				name:     "Invalid bencoded list (no end marker)",
				input:    []byte("li1ei2ei3e"),
				expected: []interface{}{},
				bytes:    0,
				err:      errors.New("invalid list format: no end marker"),
			},
			{
				name:     "Invalid bencoded list (invalid item)",
				input:    []byte("li1ei2ei3e4e"),
				expected: []interface{}{},
				bytes:    0,
				err:      errors.New("error decoding list item: invalid string format: no colon found"),
			},
		})
	})

	t.Run("Edge cases", func(t *testing.T) {
		runTests(t, []testCase{
			{
				name:     "Empty data",
				input:    []byte(""),
				expected: nil,
				bytes:    0,
				err:      errors.New("empty data"),
			},
			{
				name:     "Invalid bencoded string (non-numeric length)",
				input:    []byte("a:hello"),
				expected: nil,
				bytes:    0,
				err:      fmt.Errorf("unkown type: %c", 'a'),
			},
		})
	})
}

// testCase represents a single test scenario
type testCase struct {
	name     string
	input    []byte
	expected interface{}
	bytes    int
	err      error
}

// runTests executes a slice of test cases with consistent validation logic
func runTests(t *testing.T, tests []testCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, n, err := Decode(tt.input)

			// Check error correctness
			if err != nil && tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("Decode(%q) error = %v, want error = %v", tt.input, err, tt.err)
				}
			} else if (err == nil) != (tt.err == nil) {
				t.Errorf("Decode(%q) error = %v, want error = %v", tt.input, err, tt.err)
			}

			// Check bytes read
			if n != tt.bytes {
				t.Errorf("Decode(%q) bytes = %d, want %d", tt.input, n, tt.bytes)
			}

			// Check result, with special handling for empty lists
			if listResult, ok := result.([]interface{}); ok {
				if expectedList, ok2 := tt.expected.([]interface{}); ok2 {
					if len(expectedList) == 0 && len(listResult) == 0 {
						// Both are empty lists, this is fine
					} else if !reflect.DeepEqual(listResult, expectedList) {
						t.Errorf("Decode(%q) result = %v, want %v", tt.input, result, tt.expected)
					}
				} else if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("Decode(%q) result = %v, want %v", tt.input, result, tt.expected)
				}
			} else if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Decode(%q) result = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
