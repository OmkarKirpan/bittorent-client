package torrent

import (
	"encoding/hex"
	"testing"
)

func TestParseFromFile(t *testing.T) {
	torrentFile, err := ParseFromFile("../Debian.torrent") // Adjust the path if needed
	if err != nil {
		t.Fatalf("Failed to parse torrent file: %v", err)
	}

	t.Logf("Torrent Name: %s", torrentFile.Info.Name)
	t.Logf("Announce URL: %s", torrentFile.Announce)
	t.Logf("Piece Length: %d bytes", torrentFile.Info.PieceLength)

	if torrentFile.Info.Length > 0 {
		t.Logf("Single File Mode: %s (%d bytes)", torrentFile.Info.Name, torrentFile.Info.Length)
	} else {
		t.Logf("Multiple Files Mode: %d files", len(torrentFile.Info.Files))
		for i, file := range torrentFile.Info.Files {
			t.Logf("  File %d: %v (%d bytes)", i+1, file.Path, file.Length)
		}
	}
}

func TestInfoHash(t *testing.T) {
	// Parse the torrent file. Adjust the path if needed.
	torrentFile, err := ParseFromFile("../Debian.torrent")
	if err != nil {
		t.Fatalf("Failed to parse torrent file: %v", err)
	}

	// Compute the info hash
	hash1, err := torrentFile.InfoHash()
	if err != nil {
		t.Fatalf("InfoHash() returned error: %v", err)
	}

	// Compute the info hash a second time to ensure consistency
	hash2, err := torrentFile.InfoHash()
	if err != nil {
		t.Fatalf("InfoHash() returned error on second call: %v", err)
	}

	// Check that repeated calls produce the same result
	if hash1 != hash2 {
		t.Errorf("Inconsistent InfoHash results: first %x, second %x", hash1, hash2)
	}

	// Log the computed info hash in hexadecimal format
	computedHash := hex.EncodeToString(hash1[:])
	t.Logf("Computed InfoHash: %s", computedHash)

	expectedHash := "83e53cb48c4af4989cd1a53a5b4671da821b1ff4" // Replace with the expected SHA-1 hash in hex
	if computedHash != expectedHash {
		t.Errorf("InfoHash mismatch: got %s, expected %s", computedHash, expectedHash)
	}
}
