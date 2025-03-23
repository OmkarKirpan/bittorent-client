package peer

import (
	"bytes"
	"testing"
)

func TestHandshakeSerialization(t *testing.T) {
	infoHash := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	peerID := [20]byte{21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}

	h := NewHandshake(infoHash, peerID)

	// Serialize
	buf := h.Serialize()

	// Check length
	if len(buf) != HandshakeLength {
		t.Errorf("Expected handshake length %d, got %d", HandshakeLength, len(buf))
	}

	// Check protocol string length
	if buf[0] != byte(len(ProtocolIdentifier)) {
		t.Errorf("Expected protocol string length %d, got %d", len(ProtocolIdentifier), buf[0])
	}

	// Deserialize
	reader := bytes.NewReader(buf)
	parsed, err := ParseHandshake(reader)
	if err != nil {
		t.Fatalf("ParseHandshake failed: %v", err)
	}

	// Check fields
	if parsed.Pstr != ProtocolIdentifier {
		t.Errorf("Expected protocol identifier %s, got %s", ProtocolIdentifier, parsed.Pstr)
	}

	if !bytes.Equal(parsed.InfoHash[:], infoHash[:]) {
		t.Errorf("Info hash mismatch")
	}

	if !bytes.Equal(parsed.PeerID[:], peerID[:]) {
		t.Errorf("Peer ID mismatch")
	}
}

func TestExtensionBits(t *testing.T) {
	h := NewHandshake([20]byte{}, [20]byte{})

	// Test setting and checking extension bits
	h.SetExtension(ExtensionDHT)
	if !h.HasExtension(ExtensionDHT) {
		t.Errorf("DHT extension bit should be set")
	}

	if h.HasExtension(ExtensionFast) {
		t.Errorf("Fast extension bit should not be set")
	}

	h.SetExtension(ExtensionExtensions)
	if !h.HasExtension(ExtensionExtensions) {
		t.Errorf("Extension protocol bit should be set")
	}

	// Check raw values - DHT is bit 0 in reserved byte 7
	// This should set the rightmost bit (2^0 = 1) in byte 7
	if h.Reserved[7] != 1 {
		t.Errorf("Expected byte 7 to have value 1, got %d", h.Reserved[7])
	}

	// Extension protocol is bit 5 in reserved byte 5
	// This should set the 6th bit (2^5 = 32) in byte 5
	if h.Reserved[5] != 32 {
		t.Errorf("Expected byte 5 to have value 32, got %d", h.Reserved[5])
	}
}
