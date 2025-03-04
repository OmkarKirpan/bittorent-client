package main

import (
	"fmt"
	"log"

	"github.com/omkarkirpan/bittorrent-client/torrent"
)

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
}
