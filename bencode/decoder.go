package bencode

import (
	"errors"
	"fmt"
	"strconv"
)

// Decode parses a bencoded string into its corresponding Go type
func Decode(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("empty data")
	}

	switch data[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return decodeString(data)
	default:
		return nil, 0, fmt.Errorf("unkown type: %c", data[0])
	}
}

// decodeString parses a bencoded string
// Format: <length>:<contents>
// Example: 5:hello -> "hello"
func decodeString(data []byte) (string, int, error) {
	i := 0

	// Find the colon separator
	for i < len(data) && data[i] != ':' {
		i++
	}

	if i >= len(data) {
		return "", 0, errors.New("invalid string format: no colon found")
	}

	// Parse the length of the string
	length, err := strconv.Atoi(string(data[:i]))
	if err != nil {
		return "", 0, fmt.Errorf(("invalid string format: %v"), err)
	}

	// Check if we have enough data
	if i+1+length > len(data) {
		return "", 0, errors.New("string data too short")
	}

	// Extract the string content
	result := string(data[i+1 : i+1+length])

	// Return string, total bytes consumed, nil error
	return result, i + 1 + length, nil
}
