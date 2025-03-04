package tracker

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/omkarkirpan/bittorrent-client/bencode"
	"github.com/omkarkirpan/bittorrent-client/torrent"
)

// Peer represents a peer in the BitTorrent network
type Peer struct {
	IP   net.IP
	Port uint16
}

// String returns a string representation of a peer
func (p Peer) String() string {
	return fmt.Sprintf("%s:%d", p.IP.String(), p.Port)
}

// TrackerResponse represents the response from a tracker
type TrackerResponse struct {
	Interval    int    `bencode:"interval"`
	MinInterval int    `bencode:"min interval,omitempty"`
	Complete    int    `bencode:"complete,omitempty"`
	Incomplete  int    `bencode:"incomplete,omitempty"`
	Peers       string `bencode:"peers"`
	// We'll ignore the dictionary model of peers for now
}

// RequestPeers sends a request to the tracker and returns a list of peers
func RequestPeers(torrentFile *torrent.TorrentFile, port uint16) ([]Peer, error) {
	// Generate a random peer ID (20 bytes)
	peerId := generatePeerId()

	// Calculate the info hash
	infoHash, err := torrentFile.InfoHash()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate info hash: %v", err)
	}

	// Construct the tracker URL with query parameters
	announceURL, err := url.Parse(torrentFile.Announce)
	if err != nil {
		return nil, fmt.Errorf("invalid announce URL: %v", err)
	}

	q := announceURL.Query()
	q.Set("info_hash", string(infoHash[:]))
	q.Set("peer_id", string(peerId[:]))
	q.Set("port", strconv.Itoa(int(port)))
	q.Set("uploaded", "0")
	q.Set("downloaded", "0")
	q.Set("left", strconv.FormatInt(torrentFile.TotalLength(), 10))
	q.Set("compact", "1")
	announceURL.RawQuery = q.Encode()

	// Send the HTTP GET request to the tracker
	resp, err := http.Get(announceURL.String())
	if err != nil {
		return nil, fmt.Errorf("tracker request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tracker response: %v", err)
	}

	trackerResp, err := parseTrackerResponse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tracker response: %v", err)
	}

	// Parse the compact peer list
	peers, err := parsePeers(trackerResp.Peers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer list: %v", err)
	}

	return peers, nil
}

// generatePeerId creates a 20-byte peer ID with the prefix -GO0001-
func generatePeerId() [20]byte {
	var id [20]byte

	// Use a common prefix to identify our client
	prefix := []byte("-GO0001-")
	copy(id[:], prefix)

	// Fill the rest with random bytes
	rand.Seed(time.Now().UnixNano())
	rand.Read(id[len(prefix):])

	return id
}

// parseTrackerResponse decodes the bencoded tracker response
func parseTrackerResponse(body []byte) (*TrackerResponse, error) {
	decoded, _, err := bencode.Decode(body)
	if err != nil {
		return nil, err
	}

	dict, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("tracker response is not a dictionary")
	}

	response := &TrackerResponse{}

	// Parse required fields
	if interval, ok := dict["interval"].(int64); ok {
		response.Interval = int(interval)
	} else {
		return nil, fmt.Errorf("missing or invalid interval")
	}

	if peers, ok := dict["peers"].(string); ok {
		response.Peers = peers
	} else {
		return nil, fmt.Errorf("missing or invalid peers")
	}

	// Parse optional fields
	if minInterval, ok := dict["min interval"].(int64); ok {
		response.MinInterval = int(minInterval)
	}

	if complete, ok := dict["complete"].(int64); ok {
		response.Complete = int(complete)
	}

	if incomplete, ok := dict["incomplete"].(int64); ok {
		response.Incomplete = int(incomplete)
	}

	return response, nil
}

// parsePeers extracts peers from the compact peer list
func parsePeers(compactPeers string) ([]Peer, error) {
	peerData := []byte(compactPeers)

	// Each peer is represented by 6 bytes: 4 for IP, 2 for port
	if len(peerData)%6 != 0 {
		return nil, fmt.Errorf("invalid peer list length: %d", len(peerData))
	}

	peers := make([]Peer, 0, len(peerData)/6)

	for i := 0; i < len(peerData); i += 6 {
		ip := net.IPv4(peerData[i], peerData[i+1], peerData[i+2], peerData[i+3])
		port := binary.BigEndian.Uint16(peerData[i+4 : i+6])

		peers = append(peers, Peer{IP: ip, Port: port})
	}

	return peers, nil
}
