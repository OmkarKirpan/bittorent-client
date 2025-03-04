package bencode

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

// EncodeDict encodes a map into a bencoded dictionary
func EncodeDict(dict map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer

	// Start with 'd'
	buf.WriteByte('d')

	// Sort keys for canonical ordering (required by the BitTorrent spec)
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Encode each key-value pair
	for _, key := range keys {
		// Encode key as a bencoded string
		keyLen := strconv.Itoa(len(key))
		buf.WriteString(keyLen)
		buf.WriteByte(':')
		buf.WriteString(key)

		// Encode value based on its type
		err := encodeValue(&buf, dict[key])
		if err != nil {
			return nil, err
		}
	}

	// End with 'e'
	buf.WriteByte('e')

	return buf.Bytes(), nil
}

// encodeValue encodes a value based on its type
func encodeValue(buf *bytes.Buffer, value interface{}) error {
	switch v := value.(type) {
	case string:
		// Format: <length>:<contents>
		buf.WriteString(strconv.Itoa(len(v)))
		buf.WriteByte(':')
		buf.WriteString(v)
	case int, int64:
		// Format: i<number>e
		var intVal int64
		if i, ok := v.(int); ok {
			intVal = int64(i)
		} else {
			intVal = v.(int64)
		}
		buf.WriteByte('i')
		buf.WriteString(strconv.FormatInt(intVal, 10))
		buf.WriteByte('e')
	case []interface{}:
		// Format: l<contents>e
		buf.WriteByte('l')
		for _, item := range v {
			err := encodeValue(buf, item)
			if err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	case map[string]interface{}:
		// Encode dictionary recursively
		dictBytes, err := EncodeDict(v)
		if err != nil {
			return err
		}
		buf.Write(dictBytes)
	case []string:
		// Special case for string slices
		buf.WriteByte('l')
		for _, item := range v {
			buf.WriteString(strconv.Itoa(len(item)))
			buf.WriteByte(':')
			buf.WriteString(item)
		}
		buf.WriteByte('e')
	default:
		return fmt.Errorf("unsupported type for encoding: %T", v)
	}
	return nil
}
