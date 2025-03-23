package main

import (
	"fmt"
	"log"
	"time"

	"github.com/omkarkirpan/bittorrent-client/peer"
	"github.com/omkarkirpan/bittorrent-client/torrent"
	"github.com/omkarkirpan/bittorrent-client/tracker"
)

// humanReadableSize converts bytes to a human-readable format.
func humanReadableSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

func main() {
	fmt.Println("BitTorrent Client")

	// Test parsing a torrent file
	torrentFile, err := torrent.ParseFromFile("Debian.torrent")
	if err != nil {
		log.Fatalf("Error parsing torrent file: %v", err)
	}

	fmt.Printf("Torrent Name: %s\n", torrentFile.Info.Name)
	fmt.Printf("Announce URL: %s\n", torrentFile.Announce)
	fmt.Printf("Piece Length: %d bytes\n", torrentFile.Info.PieceLength)

	// Print file information
	if torrentFile.Info.Length > 0 {
		fmt.Printf("Single File Mode: %s (%d bytes)\n", torrentFile.Info.Name, torrentFile.Info.Length)
	} else {
		fmt.Printf("Multiple Files Mode: %d files\n", len(torrentFile.Info.Files))
		for i, file := range torrentFile.Info.Files {
			fmt.Printf("  File %d: %v (%d bytes)\n", i+1, file.Path, file.Length)
		}
	}

	// Calculate and print the info hash
	infoHash, err := torrentFile.InfoHash()
	if err != nil {
		log.Fatalf("Error calculating info hash: %v", err)
	}

	fmt.Printf("Info Hash: %x\n", infoHash)

	// Print information about pieces
	numPieces := torrentFile.NumPieces()
	fmt.Printf("Number of Pieces: %d\n", numPieces)
	fmt.Printf("Total Size: %s bytes\n", humanReadableSize(torrentFile.TotalLength()))

	// Print first few piece hashes
	displayCount := 3
	if displayCount > numPieces {
		displayCount = numPieces
	}

	for i := 0; i < displayCount; i++ {
		hash, err := torrentFile.PieceHash(i)
		if err != nil {
			log.Fatalf("Error getting piece hash: %v", err)
		}
		fmt.Printf("Piece %d Hash: %x (Length: %s bytes)\n", i, hash, humanReadableSize(torrentFile.PieceLength(i)))
	}

	// Also print last piece if there are more than the display count
	if numPieces > displayCount {
		hash, err := torrentFile.PieceHash(numPieces - 1)
		if err != nil {
			log.Fatalf("Error getting piece hash: %v", err)
		}
		fmt.Printf("Last Piece (%d) Hash: %x (Length: %s bytes)\n",
			numPieces-1, hash, humanReadableSize(torrentFile.PieceLength(numPieces-1)))
	}

	// Discover peers
	fmt.Println("\nDiscovering peers...")
	peers, err := tracker.RequestPeers(torrentFile, 6881) // 6881 is a common BitTorrent port
	if err != nil {
		log.Fatalf("Error discovering peers: %v", err)
	}

	fmt.Printf("Found %d peers:\n", len(peers))
	for i, p := range peers {
		if i >= 5 {
			fmt.Printf("... and %d more\n", len(peers)-5)
			break
		}
		fmt.Printf("  %s\n", p.String())
	}

	// Generate a peer ID (this should match the one used in tracker request)
	var peerId [20]byte
	copy(peerId[:], []byte("-GO0001-1234567890123"))

	// Test peer handshake with the first few peers
	fmt.Println("\nAttempting handshakes with peers...")

	maxPeersToTry := 5
	if len(peers) < maxPeersToTry {
		maxPeersToTry = len(peers)
	}

	handshakeSuccessful := false
	var successfulPeer tracker.Peer
	var successfulHandshake *peer.Handshake

	// Try each peer until we get a successful handshake
	for i := 0; i < maxPeersToTry && !handshakeSuccessful; i++ {
		fmt.Printf("Trying peer %d: %s\n", i+1, peers[i].String())

		// Set a timeout for the handshake
		handshakeChan := make(chan struct {
			handshake *peer.Handshake
			err       error
		})

		go func(p tracker.Peer) {
			handshake, conn, err := peer.PerformHandshake(p.String(), infoHash, peerId)
			if err == nil && conn != nil {
				defer conn.Close()
			}
			handshakeChan <- struct {
				handshake *peer.Handshake
				err       error
			}{handshake, err}
		}(peers[i])

		// Wait for handshake or timeout
		select {
		case result := <-handshakeChan:
			if result.err != nil {
				fmt.Printf("  Handshake failed: %v\n", result.err)
			} else {
				fmt.Println("  Handshake successful!")
				handshakeSuccessful = true
				successfulPeer = peers[i]
				successfulHandshake = result.handshake
			}
		case <-time.After(5 * time.Second):
			fmt.Println("  Handshake timed out")
		}
	}

	// If we found a successful peer, display information about it
	if handshakeSuccessful {
		fmt.Printf("\nSuccessfully connected to peer: %s\n", successfulPeer.String())
		fmt.Printf("Remote peer ID: %x\n", successfulHandshake.PeerID)

		// Check for extension support
		if successfulHandshake.HasExtension(peer.ExtensionDHT) {
			fmt.Println("Peer supports DHT")
		}
		if successfulHandshake.HasExtension(peer.ExtensionExtensions) {
			fmt.Println("Peer supports Extension Protocol")
		}
		if successfulHandshake.HasExtension(peer.ExtensionFast) {
			fmt.Println("Peer supports Fast Extension")
		}
	} else {
		fmt.Println("\nFailed to handshake with any peers. This can happen if:")
		fmt.Println("1. The peers are not online or are not accepting connections")
		fmt.Println("2. Network restrictions are preventing the connections")
		fmt.Println("3. The peers have reached their connection limit")
	}
}
