package peer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// MessageType represents the type of BitTorrent message
type MessageType uint8

// Message types as defined in the BitTorrent protocol
const (
	MsgChoke         MessageType = 0
	MsgUnchoke       MessageType = 1
	MsgInterested    MessageType = 2
	MsgNotInterested MessageType = 3
	MsgHave          MessageType = 4
	MsgBitfield      MessageType = 5
	MsgRequest       MessageType = 6
	MsgPiece         MessageType = 7
	MsgCancel        MessageType = 8
	MsgPort          MessageType = 9 // Used by DHT extension
)

// Message represents a BitTorrent protocol message
type Message struct {
	Length  uint32
	Type    MessageType
	Payload []byte
}

// KeepAliveMessage is a message with a zero length and no ID or payload
var KeepAliveMessage = Message{Length: 0, Type: 0, Payload: nil}

// Serialize converts a Message to its wire format
func (m *Message) Serialize() []byte {
	if m.Length == 0 {
		// Keep-alive message: just 4 bytes of zero
		return make([]byte, 4)
	}

	// Message length (4 bytes) + message type (1 byte) + payload
	buffer := make([]byte, 4+1+len(m.Payload))

	// Set the message length (excluding the length field itself)
	binary.BigEndian.PutUint32(buffer[0:4], m.Length)

	// Set the message type
	buffer[4] = byte(m.Type)

	// Copy the payload if any
	if len(m.Payload) > 0 {
		copy(buffer[5:], m.Payload)
	}

	return buffer
}

// ReadMessage reads a message from an io.Reader
func ReadMessage(r io.Reader) (*Message, error) {
	// Read message length (4 bytes)
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	// If length is 0, it's a keep-alive message
	if length == 0 {
		return &KeepAliveMessage, nil
	}

	// Read message type and payload
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	// Message type is the first byte of the message
	messageType := MessageType(messageBuf[0])

	// Payload is the rest of the message
	var payload []byte
	if length > 1 {
		payload = messageBuf[1:]
	}

	return &Message{
		Length:  length,
		Type:    messageType,
		Payload: payload,
	}, nil
}

// FormatMessage creates a message of the specified type with the given payload
func FormatMessage(messageType MessageType, payload []byte) *Message {
	length := uint32(1) // 1 byte for message type
	if payload != nil {
		length += uint32(len(payload))
	}

	return &Message{
		Length:  length,
		Type:    messageType,
		Payload: payload,
	}
}

// RequestMessage creates a request message for a piece
func RequestMessage(index, begin, length uint32) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], index)   // Piece index
	binary.BigEndian.PutUint32(payload[4:8], begin)   // Offset within the piece
	binary.BigEndian.PutUint32(payload[8:12], length) // Length of the requested block

	return FormatMessage(MsgRequest, payload)
}

// ParseHave parses a HAVE message payload
func ParseHave(msg *Message) (uint32, error) {
	if msg.Type != MsgHave {
		return 0, errors.New("not a HAVE message")
	}

	if len(msg.Payload) != 4 {
		return 0, errors.New("invalid HAVE message payload length")
	}

	return binary.BigEndian.Uint32(msg.Payload), nil
}

// ParsePiece parses a PIECE message payload
func ParsePiece(pieceIndex uint32, msg *Message) (uint32, []byte, error) {
	if msg.Type != MsgPiece {
		return 0, nil, errors.New("not a PIECE message")
	}

	if len(msg.Payload) < 8 {
		return 0, nil, errors.New("invalid PIECE message payload length")
	}

	index := binary.BigEndian.Uint32(msg.Payload[0:4])
	if index != pieceIndex {
		return 0, nil, fmt.Errorf("unexpected piece index: expected %d, got %d", pieceIndex, index)
	}

	begin := binary.BigEndian.Uint32(msg.Payload[4:8])
	data := msg.Payload[8:]

	return begin, data, nil
}

// String returns a string representation of a message
func (m *Message) String() string {
	if m.Length == 0 {
		return "KeepAlive"
	}

	var typeName string
	switch m.Type {
	case MsgChoke:
		typeName = "Choke"
	case MsgUnchoke:
		typeName = "Unchoke"
	case MsgInterested:
		typeName = "Interested"
	case MsgNotInterested:
		typeName = "NotInterested"
	case MsgHave:
		index := binary.BigEndian.Uint32(m.Payload)
		return fmt.Sprintf("Have[%d]", index)
	case MsgBitfield:
		return fmt.Sprintf("Bitfield[%d bytes]", len(m.Payload))
	case MsgRequest:
		index := binary.BigEndian.Uint32(m.Payload[0:4])
		begin := binary.BigEndian.Uint32(m.Payload[4:8])
		length := binary.BigEndian.Uint32(m.Payload[8:12])
		return fmt.Sprintf("Request[%d:%d:%d]", index, begin, length)
	case MsgPiece:
		index := binary.BigEndian.Uint32(m.Payload[0:4])
		begin := binary.BigEndian.Uint32(m.Payload[4:8])
		return fmt.Sprintf("Piece[%d:%d:%d bytes]", index, begin, len(m.Payload)-8)
	case MsgCancel:
		index := binary.BigEndian.Uint32(m.Payload[0:4])
		begin := binary.BigEndian.Uint32(m.Payload[4:8])
		length := binary.BigEndian.Uint32(m.Payload[8:12])
		return fmt.Sprintf("Cancel[%d:%d:%d]", index, begin, length)
	case MsgPort:
		port := binary.BigEndian.Uint16(m.Payload)
		return fmt.Sprintf("Port[%d]", port)
	default:
		typeName = fmt.Sprintf("Unknown(%d)", m.Type)
	}

	if len(m.Payload) > 0 {
		return fmt.Sprintf("%s[%d bytes]", typeName, len(m.Payload))
	}

	return typeName
}
