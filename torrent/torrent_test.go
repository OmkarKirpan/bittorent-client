package torrent

import (
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
