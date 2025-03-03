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
	case 'i':
		return decodeInteger(data)
	case 'l':
		return decodeList(data)
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

// decodeInteger parses a bencoded integer
// Format: i<number>e
// Example: i42e -> 42
func decodeInteger(data []byte) (int64, int, error) {
	if len(data) < 2 || data[0] != 'i' {
		return 0, 0, errors.New("invalid integer format")
	}

	// Find the end marker 'e'
	endIndex := 1
	for endIndex < len(data) && data[endIndex] != 'e' {
		endIndex++
	}

	if endIndex >= len(data) {
		return 0, 0, errors.New("invalid integer format: no end marker")
	}

	// Parse the integer
	numStr := string(data[1:endIndex])

	// Check for leading zeros or empty string
	if len(numStr) > 1 && numStr[0] == '0' {
		return 0, 0, errors.New("invalid integer format: leading zeros")
	}

	// Check for negative zero
	if len(numStr) > 1 && numStr[0] == '-' && numStr[1] == '0' {
		return 0, 0, errors.New("invalid integer format: negative zero")
	}

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid integer: %v", err)
	}

	// Return value, total bytes consumed, nil error
	return num, endIndex + 1, nil
}

// decodeList parses a bencoded list
// Format: l<contents>e
// Example: li1ei2ei3ee -> [1, 2, 3]
func decodeList(data []byte) ([]interface{}, int, error) {
	if len(data) < 2 || data[0] != 'l' {
		return nil, 0, errors.New("invalid list format")
	}

	result := []interface{}{}
	pos := 1 // Skip the 'l' marker

	for pos < len(data) && data[pos] != 'e' {
		// Decode the next item in the list
		item, bytesRead, err := Decode(data[pos:])
		if err != nil {
			return nil, 0, fmt.Errorf("error decoding list item: %v", err)
		}

		// Add item to result and move position forward
		result = append(result, item)
		pos += bytesRead
	}

	if pos >= len(data) {
		return nil, 0, errors.New("invalid list format: no end marker")
	}

	// Skip the 'e' marker
	pos++

	// Return list, total bytes consumed, nil error
	return result, pos, nil
}
