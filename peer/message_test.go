package peer

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestMessageSerialization(t *testing.T) {
	// Test keep-alive message
	keepAlive := &Message{Length: 0, Type: 0, Payload: nil}
	serialized := keepAlive.Serialize()

	if len(serialized) != 4 {
		t.Errorf("Expected keep-alive message length 4, got %d", len(serialized))
	}

	// All bytes should be zero
	for _, b := range serialized {
		if b != 0 {
			t.Errorf("Expected keep-alive message byte to be 0, got %d", b)
		}
	}

	// Test message with payload
	payload := []byte{1, 2, 3, 4}
	msg := FormatMessage(MsgRequest, payload)
	serialized = msg.Serialize()

	// 4 bytes length + 1 byte type + 4 bytes payload = 9 bytes
	if len(serialized) != 9 {
		t.Errorf("Expected message length 9, got %d", len(serialized))
	}

	// Length should be 5 (1 byte type + 4 bytes payload)
	length := binary.BigEndian.Uint32(serialized[0:4])
	if length != 5 {
		t.Errorf("Expected message length field 5, got %d", length)
	}

	// Type should be 6 (Request)
	if serialized[4] != byte(MsgRequest) {
		t.Errorf("Expected message type %d, got %d", MsgRequest, serialized[4])
	}

	// Payload should match
	if !bytes.Equal(serialized[5:], payload) {
		t.Errorf("Payload mismatch")
	}
}

func TestMessageDeserialization(t *testing.T) {
	// Create a keep-alive message
	keepAliveBytes := make([]byte, 4)
	reader := bytes.NewReader(keepAliveBytes)

	// Read it back
	msg, err := ReadMessage(reader)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	// Check fields
	if msg.Length != 0 {
		t.Errorf("Expected length 0, got %d", msg.Length)
	}

	// Create a message with payload
	msgBytes := make([]byte, 9)
	binary.BigEndian.PutUint32(msgBytes[0:4], 5)  // Length = 5 (1 for type + 4 for payload)
	msgBytes[4] = byte(MsgHave)                   // Type = Have
	binary.BigEndian.PutUint32(msgBytes[5:9], 42) // Payload = piece index 42

	reader = bytes.NewReader(msgBytes)

	// Read it back
	msg, err = ReadMessage(reader)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	// Check fields
	if msg.Length != 5 {
		t.Errorf("Expected length 5, got %d", msg.Length)
	}

	if msg.Type != MsgHave {
		t.Errorf("Expected type Have (%d), got %d", MsgHave, msg.Type)
	}

	// Parse the HAVE message
	index, err := ParseHave(msg)
	if err != nil {
		t.Fatalf("ParseHave failed: %v", err)
	}

	if index != 42 {
		t.Errorf("Expected piece index 42, got %d", index)
	}
}

func TestMessageHelpers(t *testing.T) {
	// Test RequestMessage
	index := uint32(7)
	begin := uint32(1024)
	length := uint32(16384)

	requestMsg := RequestMessage(index, begin, length)

	if requestMsg.Type != MsgRequest {
		t.Errorf("Expected message type Request (%d), got %d", MsgRequest, requestMsg.Type)
	}

	if len(requestMsg.Payload) != 12 {
		t.Errorf("Expected payload length 12, got %d", len(requestMsg.Payload))
	}

	// Check payload values
	indexValue := binary.BigEndian.Uint32(requestMsg.Payload[0:4])
	beginValue := binary.BigEndian.Uint32(requestMsg.Payload[4:8])
	lengthValue := binary.BigEndian.Uint32(requestMsg.Payload[8:12])

	if indexValue != index {
		t.Errorf("Expected index %d, got %d", index, indexValue)
	}

	if beginValue != begin {
		t.Errorf("Expected begin %d, got %d", begin, beginValue)
	}

	if lengthValue != length {
		t.Errorf("Expected length %d, got %d", length, lengthValue)
	}

	// Test message string representation
	msgString := requestMsg.String()
	expected := "Request[7:1024:16384]"
	if msgString != expected {
		t.Errorf("Expected string representation %s, got %s", expected, msgString)
	}
}
