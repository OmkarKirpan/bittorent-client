package torrent

import (
	"encoding/hex"
	"strconv"
	"testing"
)

// loadTorrentFile is a helper function to parse the torrent file and fail the test if it cannot be loaded.
func loadTorrentFile(t *testing.T) *TorrentFile {
	t.Helper()
	torrentFile, err := ParseFromFile("../Debian.torrent") // Adjust the path if needed
	if err != nil {
		t.Fatalf("Failed to parse torrent file: %v", err)
	}
	return torrentFile
}

func TestParseFromFile(t *testing.T) {
	torrentFile := loadTorrentFile(t)

	t.Run("BasicInfo", func(t *testing.T) {
		t.Logf("Torrent Name: %s", torrentFile.Info.Name)
		t.Logf("Announce URL: %s", torrentFile.Announce)
		t.Logf("Piece Length: %d bytes", torrentFile.Info.PieceLength)
	})

	t.Run("FileMode", func(t *testing.T) {
		if torrentFile.Info.Length > 0 {
			t.Logf("Single File Mode: %s (%d bytes)", torrentFile.Info.Name, torrentFile.Info.Length)
		} else {
			t.Logf("Multiple Files Mode: %d files", len(torrentFile.Info.Files))
			for i, file := range torrentFile.Info.Files {
				t.Logf("  File %d: %v (%d bytes)", i+1, file.Path, file.Length)
			}
		}
	})
}

func TestInfoHash(t *testing.T) {
	torrentFile := loadTorrentFile(t)

	t.Run("ConsistencyCheck", func(t *testing.T) {
		hash1, err := torrentFile.InfoHash()
		if err != nil {
			t.Fatalf("InfoHash() returned error: %v", err)
		}
		hash2, err := torrentFile.InfoHash()
		if err != nil {
			t.Fatalf("InfoHash() returned error on second call: %v", err)
		}

		if hash1 != hash2 {
			t.Errorf("Inconsistent InfoHash results: first %x, second %x", hash1, hash2)
		}
	})

	t.Run("ExpectedHash", func(t *testing.T) {
		hash, err := torrentFile.InfoHash()
		if err != nil {
			t.Fatalf("InfoHash() returned error: %v", err)
		}
		computedHash := hex.EncodeToString(hash[:])
		t.Logf("Computed InfoHash: %s", computedHash)

		expectedHash := "83e53cb48c4af4989cd1a53a5b4671da821b1ff4" // Replace with the expected SHA-1 hash in hex
		if computedHash != expectedHash {
			t.Errorf("InfoHash mismatch: got %s, expected %s", computedHash, expectedHash)
		}
	})
}

func TestPieceHash(t *testing.T) {
	torrentFile := loadTorrentFile(t)

	// Table-driven tests for piece hashes.
	testCases := []struct {
		pieceIndex int
		expected   string
	}{
		{pieceIndex: 0, expected: "c8e809c29620b1dc0c90727a038b375ce6dbbc58"},
		{pieceIndex: 1, expected: "4e9aa6377b79c31d94001c1e39dbe624fcd0fbef"},
		{pieceIndex: 2, expected: "4f9dd1ea176d4de11affe8cf8c9aef2c96f205f5"},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run("PieceIndex_"+strconv.Itoa(tc.pieceIndex), func(t *testing.T) {
			hash, err := torrentFile.PieceHash(tc.pieceIndex)
			if err != nil {
				t.Fatalf("Failed to compute hash for piece %d: %v", tc.pieceIndex, err)
			}

			computedHash := hex.EncodeToString(hash[:])
			if computedHash != tc.expected {
				t.Errorf("Piece %d hash mismatch: got %s, expected %s", tc.pieceIndex, computedHash, tc.expected)
			} else {
				t.Logf("Piece %d hash verified: %s", tc.pieceIndex, computedHash)
			}
		})
	}
}
