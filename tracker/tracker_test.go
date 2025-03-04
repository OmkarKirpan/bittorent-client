package tracker_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/omkarkirpan/bittorrent-client/torrent"
	"github.com/omkarkirpan/bittorrent-client/tracker"
)

// TestRequestPeersSuccess simulates a tracker response with a compact peer string.
func TestRequestPeersSuccess(t *testing.T) {
	// Build the compact peer string.
	// First peer: IP: 127.0.0.1, Port: 6881 (0x1ae1)
	// Second peer: IP: 192.168.0.1, Port: 6881 (0x1ae1)
	compactPeers := []byte{
		0x7f, 0x00, 0x00, 0x01, 0x1a, 0xe1,
		0xc0, 0xa8, 0x00, 0x01, 0x1a, 0xe1,
	}
	// Construct the bencoded response.
	// The bencode string represents a dictionary with two keys: interval and peers.
	// "12:" indicates that the following 12 bytes form the peer string.
	response := "d8:intervali1800e5:peers12:" + string(compactPeers) + "e"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write the bencoded response.
		w.Write([]byte(response))
	}))
	defer ts.Close()

	// Create a dummy torrent file with the tracker announce URL set to our test server.
	torrentFile := &torrent.TorrentFile{
		Announce: ts.URL,
		Info: torrent.TorrentInfo{
			Name:        "dummy",
			PieceLength: 262144,
		},
	}

	// Call RequestPeers using the dummy torrent file.
	peers, err := tracker.RequestPeers(torrentFile, 6881)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify that we received two peers.
	expectedCount := 2
	if len(peers) != expectedCount {
		t.Errorf("Expected %d peers, got %d", expectedCount, len(peers))
	}

	// Verify each peer's IP and port.
	// Adjust the field access according to your Peer struct implementation.
	peer0 := peers[0]
	peer1 := peers[1]

	// First peer should be 127.0.0.1:6881.
	if peer0.IP.String() != "127.0.0.1" || peer0.Port != 6881 {
		t.Errorf("Unexpected peer 0: got %s:%d, expected 127.0.0.1:6881", peer0.IP.String(), peer0.Port)
	}

	// Second peer should be 192.168.0.1:6881.
	if peer1.IP.String() != "192.168.0.1" || peer1.Port != 6881 {
		t.Errorf("Unexpected peer 1: got %s:%d, expected 192.168.0.1:6881", peer1.IP.String(), peer1.Port)
	}
}

// TestRequestPeersError simulates an error response from the tracker.
func TestRequestPeersError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}))
	defer ts.Close()

	torrentFile := &torrent.TorrentFile{
		Announce: ts.URL,
		Info: torrent.TorrentInfo{
			Name:        "dummy",
			PieceLength: 262144,
		},
	}

	_, err := tracker.RequestPeers(torrentFile, 6881)
	if err == nil {
		t.Fatal("Expected error from RequestPeers, got nil")
	}
	t.Logf("RequestPeers returned expected error: %v", err)
}
