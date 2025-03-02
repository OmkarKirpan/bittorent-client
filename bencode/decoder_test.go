package bencode

import (
	"errors"
	"fmt"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected interface{}
		bytes    int
		err      error
	}{
		{
			name:     "Empty data",
			input:    []byte(""),
			expected: nil,
			bytes:    0,
			err:      errors.New("empty data"),
		},
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
			name:     "Invalid bencoded string (non-numeric length)",
			input:    []byte("a:hello"),
			expected: nil,
			bytes:    0,
			err:      fmt.Errorf("unkown type: %c", 'a'),
		},
		{
			name:     "String data too short",
			input:    []byte("5:hel"),
			expected: "",
			bytes:    0,
			err:      errors.New("string data too short"),
		},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, n, err := Decode(tt.input)
			if result != tt.expected || n != tt.bytes || (err != nil && err.Error() != tt.err.Error()) {
				t.Errorf("Decode(%q) = (%v, %d, %v), want (%v, %d, %v)", tt.input, result, n, err, tt.expected, tt.bytes, tt.err)
			}
		})
	}
}
