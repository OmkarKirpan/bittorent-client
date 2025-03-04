package main

import (
	"fmt"
	"log"

	"github.com/omkarkirpan/bittorrent-client/torrent"
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
}
