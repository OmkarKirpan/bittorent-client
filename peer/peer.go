// **Peer handshake** - Establish connections with peers.

package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// Constants for the protocol
const (
	ProtocolIdentifier = "BitTorrent protocol"
	HandshakeLength    = 68 // Total length of a handshake message
	ConnectionTimeout  = 3 * time.Second
)

// Handshake represents a BitTorrent handshake message
type Handshake struct {
	Pstr     string   // Protocol identifier
	Reserved [8]byte  // Reserved bytes for extensions
	InfoHash [20]byte // Torrent info hash
	PeerID   [20]byte // Peer ID
}

// Serialize converts a handshake struct to its byte representation
func (h *Handshake) Serialize() []byte {
	buf := make([]byte, HandshakeLength)

	// First byte is the length of the protocol string
	buf[0] = byte(len(h.Pstr))

	// Copy the protocol string
	copy(buf[1:], h.Pstr)

	// Copy reserved bytes
	copy(buf[1+len(h.Pstr):], h.Reserved[:])

	// Copy info hash
	copy(buf[1+len(h.Pstr)+8:], h.InfoHash[:])

	// Copy peer ID
	copy(buf[1+len(h.Pstr)+8+20:], h.PeerID[:])

	return buf
}

// ParseHandshake reads a handshake message from an io.Reader
func ParseHandshake(r io.Reader) (*Handshake, error) {
	// First read the protocol string length
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}

	pstrLen := int(lengthBuf[0])
	if pstrLen == 0 {
		return nil, errors.New("protocol string length is 0")
	}

	// Now read the rest of the handshake
	handshakeBuf := make([]byte, pstrLen+48) // pstrLen + 8 bytes reserved + 20 bytes info hash + 20 bytes peer ID
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	var h Handshake

	// Extract protocol string
	h.Pstr = string(handshakeBuf[:pstrLen])

	// Extract reserved bytes
	copy(h.Reserved[:], handshakeBuf[pstrLen:pstrLen+8])

	// Extract info hash
	copy(h.InfoHash[:], handshakeBuf[pstrLen+8:pstrLen+28])

	// Extract peer ID
	copy(h.PeerID[:], handshakeBuf[pstrLen+28:])

	return &h, nil
}

// NewHandshake creates a new handshake message
func NewHandshake(infoHash [20]byte, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     ProtocolIdentifier,
		Reserved: [8]byte{}, // All zeros by default
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

// PerformHandshake connects to a peer and completes the handshake
func PerformHandshake(peerAddr string, infoHash [20]byte, peerID [20]byte) (*Handshake, net.Conn, error) {
	conn, err := net.DialTimeout("tcp", peerAddr, ConnectionTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to peer: %v", err)
	}

	// Set deadlines to prevent hanging
	conn.SetDeadline(time.Now().Add(ConnectionTimeout))
	defer conn.SetDeadline(time.Time{}) // Reset deadline after handshake

	// Create and send our handshake
	outHandshake := NewHandshake(infoHash, peerID)
	_, err = conn.Write(outHandshake.Serialize())
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to send handshake: %v", err)
	}

	// Read and parse the response handshake
	inHandshake, err := ParseHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to read handshake: %v", err)
	}

	// Verify the info hash
	if !bytes.Equal(inHandshake.InfoHash[:], infoHash[:]) {
		conn.Close()
		return nil, nil, errors.New("info hash mismatch")
	}

	return inHandshake, conn, nil
}

// ExtensionBit represents a protocol extension bit position
type ExtensionBit uint8

const (
	// ExtensionDHT is bit 0 of reserved byte 7 (DHT protocol)
	ExtensionDHT ExtensionBit = 0

	// ExtensionExtensions is bit 5 of reserved byte 5 (BEP 10: Extension Protocol)
	ExtensionExtensions ExtensionBit = 5

	// ExtensionFast is bit 7 of reserved byte 7 (BEP 6: Fast Extension)
	ExtensionFast ExtensionBit = 7
)

// SetExtension enables a specific extension in the handshake
func (h *Handshake) SetExtension(bit ExtensionBit) {
	if bit == ExtensionDHT {
		// DHT is bit 0 of byte 7
		h.Reserved[7] |= 1
	} else if bit == ExtensionExtensions {
		// Extension Protocol is bit 5 of byte 5
		h.Reserved[5] |= 32 // 2^5 = 32
	} else if bit == ExtensionFast {
		// Fast Extension is bit 7 of byte 7
		h.Reserved[7] |= 128 // 2^7 = 128
	}
}

// HasExtension checks if a specific extension is enabled in the handshake
func (h *Handshake) HasExtension(bit ExtensionBit) bool {
	if bit == ExtensionDHT {
		return (h.Reserved[7] & 1) != 0
	} else if bit == ExtensionExtensions {
		return (h.Reserved[5] & 32) != 0
	} else if bit == ExtensionFast {
		return (h.Reserved[7] & 128) != 0
	}
	return false
}
