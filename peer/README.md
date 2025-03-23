# Peer Package

This package implements the BitTorrent peer protocol, focusing on:

1. **Peer Handshake**: Establishing connections with peers
2. **BitTorrent Protocol Messages**: Implementing the wire protocol for peer communication

## Handshake Protocol

The BitTorrent handshake is the first message exchanged between peers and follows this format:

```
<pstrlen><pstr><reserved><info_hash><peer_id>
```

Where:

- `pstrlen`: length of the protocol string (1 byte)
- `pstr`: protocol string, "BitTorrent protocol" (19 bytes)
- `reserved`: 8 reserved bytes for extensions
- `info_hash`: SHA-1 hash of the info dictionary (20 bytes)
- `peer_id`: unique identifier of the client (20 bytes)

Total handshake size: 68 bytes

## Extension Bits

The 8 reserved bytes enable BitTorrent protocol extensions:

- Byte 7, bit 0 (0x01): DHT Protocol
- Byte 5, bit 5 (0x20): Extension Protocol (BEP 10)
- Byte 7, bit 7 (0x80): Fast Extension (BEP 6)

## BitTorrent Messages

After handshake, peers communicate using messages with this format:

```
<length prefix><message ID><payload>
```

Where:

- `length prefix`: message length (4 bytes)
- `message ID`: message type (1 byte)
- `payload`: message data (variable length)

Zero-length messages are keep-alive messages.

## Message Types

- `0`: Choke
- `1`: Unchoke
- `2`: Interested
- `3`: Not Interested
- `4`: Have
- `5`: Bitfield
- `6`: Request
- `7`: Piece
- `8`: Cancel
- `9`: Port (DHT)

## Usage Example

```go
// Create a handshake
infoHash := [20]byte{...} // From torrent file
peerID := [20]byte{...}   // Your client ID
handshake := peer.NewHandshake(infoHash, peerID)

// Enable extensions
handshake.SetExtension(peer.ExtensionDHT)

// Perform handshake with peer
remoteHandshake, conn, err := peer.PerformHandshake("peer-ip:port", infoHash, peerID)
if err != nil {
    // Handle error
}
defer conn.Close()

// Check peer extensions
if remoteHandshake.HasExtension(peer.ExtensionDHT) {
    // Peer supports DHT
}

// Send a message
requestMsg := peer.RequestMessage(pieceIndex, begin, length)
conn.Write(requestMsg.Serialize())

// Read a message
msg, err := peer.ReadMessage(conn)
if err != nil {
    // Handle error
}
```
