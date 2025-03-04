package torrent

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/omkarkirpan/bittorrent-client/bencode"
)

// Encoder interface for a future bencode encoder
type Encoder interface {
	Encode(v interface{}) ([]byte, error)
}

// FileInfo represents information about a file in the torrent
type FileInfo struct {
	Length int64
	Path   []string
}

// TorrentInfo represents the "info" dictionary in a torrent file
type TorrentInfo struct {
	PieceLength int64      `bencode:"piece length"`
	Pieces      string     `bencode:"pieces"`
	Name        string     `bencode:"name"`
	Length      int64      `bencode:"length,omitempty"`
	Files       []FileInfo `bencode:"files,omitempty"`
	Private     int64      `bencode:"private,omitempty"`
}

// TorrentFile represents the structure of a torrent file
type TorrentFile struct {
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list,omitempty"`
	CreationDate int64       `bencode:"creation date,omitempty"`
	Comment      string      `bencode:"comment,omitempty"`
	CreatedBy    string      `bencode:"created by,omitempty"`
	Encoding     string      `bencode:"encoding,omitempty"`
	Info         TorrentInfo `bencode:"info"`
}

// ParseFromFile loads and parses a .torrent file
func ParseFromFile(path string) (*TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return Parse(data)
}

// Parse parses torrent data from a byte slice
func Parse(data []byte) (*TorrentFile, error) {
	decoded, _, err := bencode.Decode(data)
	if err != nil {
		return nil, err
	}

	dict, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, errors.New("torrent file is not a dictionary")
	}

	// Convert the generic map to our TorrentFile struct
	torrent := &TorrentFile{}

	// Parse announce URL
	if announce, ok := dict["announce"].(string); ok {
		torrent.Announce = announce
	} else {
		return nil, errors.New("missing or invalid announce URL")
	}

	// Parse announce-list if it exists
	if announceList, ok := dict["announce-list"].([]interface{}); ok {
		for _, tier := range announceList {
			if tierList, ok := tier.([]interface{}); ok {
				var stringTier []string
				for _, url := range tierList {
					if strURL, ok := url.(string); ok {
						stringTier = append(stringTier, strURL)
					}
				}
				torrent.AnnounceList = append(torrent.AnnounceList, stringTier)
			}
		}
	}

	// Parse optional fields
	if creationDate, ok := dict["creation date"].(int64); ok {
		torrent.CreationDate = creationDate
	}

	if comment, ok := dict["comment"].(string); ok {
		torrent.Comment = comment
	}

	if createdBy, ok := dict["created by"].(string); ok {
		torrent.CreatedBy = createdBy
	}

	if encoding, ok := dict["encoding"].(string); ok {
		torrent.Encoding = encoding
	}

	// Parse info dictionary (required)
	infoDict, ok := dict["info"].(map[string]interface{})
	if !ok {
		return nil, errors.New("missing or invalid info dictionary")
	}

	// Parse piece length (required)
	pieceLength, ok := infoDict["piece length"].(int64)
	if !ok {
		return nil, errors.New("missing or invalid piece length")
	}
	torrent.Info.PieceLength = pieceLength

	// Parse pieces (required)
	pieces, ok := infoDict["pieces"].(string)
	if !ok {
		return nil, errors.New("missing or invalid pieces")
	}
	torrent.Info.Pieces = pieces

	// Parse name (required)
	name, ok := infoDict["name"].(string)
	if !ok {
		return nil, errors.New("missing or invalid name")
	}
	torrent.Info.Name = name

	// Parse length or files (mutually exclusive)
	if length, ok := infoDict["length"].(int64); ok {
		// Single file mode
		torrent.Info.Length = length
	} else if files, ok := infoDict["files"].([]interface{}); ok {
		// Multiple files mode
		for _, fileDict := range files {
			if fileMap, ok := fileDict.(map[string]interface{}); ok {
				fileInfo := FileInfo{}

				// Parse file length
				if fileLength, ok := fileMap["length"].(int64); ok {
					fileInfo.Length = fileLength
				} else {
					return nil, errors.New("missing or invalid file length")
				}

				// Parse file path
				if pathList, ok := fileMap["path"].([]interface{}); ok {
					for _, pathElem := range pathList {
						if pathStr, ok := pathElem.(string); ok {
							fileInfo.Path = append(fileInfo.Path, pathStr)
						}
					}
				} else {
					return nil, errors.New("missing or invalid file path")
				}

				torrent.Info.Files = append(torrent.Info.Files, fileInfo)
			}
		}
	} else {
		return nil, errors.New("torrent must have either length or files")
	}

	// Parse private flag (optional)
	if private, ok := infoDict["private"].(int64); ok {
		torrent.Info.Private = private
	}

	return torrent, nil
}

// InfoHash returns the SHA-1 hash of the bencoded info dictionary
func (t *TorrentFile) InfoHash() ([20]byte, error) {
	// We need to re-encode just the info dictionary
	infoDict := map[string]interface{}{
		"piece length": t.Info.PieceLength,
		"pieces":       t.Info.Pieces,
		"name":         t.Info.Name,
	}

	// Add conditional fields
	if t.Info.Length > 0 {
		infoDict["length"] = t.Info.Length
	} else {
		// For multi-file torrents
		files := make([]interface{}, 0, len(t.Info.Files))
		for _, file := range t.Info.Files {
			fileDict := map[string]interface{}{
				"length": file.Length,
				"path":   file.Path,
			}
			files = append(files, fileDict)
		}
		infoDict["files"] = files
	}
	if t.Info.Private != 0 {
		infoDict["private"] = t.Info.Private
	}

	// For now, we'll re-encode manually since we haven't implemented an encoder yet
	encoded, err := bencode.EncodeDict(infoDict)
	if err != nil {
		return [20]byte{}, err
	}

	// Calculate SHA-1 hash
	return sha1.Sum(encoded), nil
}

// PieceHash returns the hash for a specific piece
func (t *TorrentFile) PieceHash(index int) ([20]byte, error) {
	if len(t.Info.Pieces)%20 != 0 {
		return [20]byte{}, errors.New("pieces length is not a multiple of 20")
	}

	numPieces := len(t.Info.Pieces) / 20
	if index < 0 || index >= numPieces {
		return [20]byte{}, fmt.Errorf("piece index out of range: %d (total: %d)", index, numPieces)
	}

	// Extract the 20-byte hash at the given index
	var hash [20]byte
	copy(hash[:], t.Info.Pieces[index*20:(index+1)*20])

	return hash, nil
}

// NumPieces returns the total number of pieces
func (t *TorrentFile) NumPieces() int {
	return len(t.Info.Pieces) / 20
}

// PieceLength returns the length of a piece at the given index
func (t *TorrentFile) PieceLength(index int) int64 {
	if index < 0 || index >= t.NumPieces() {
		return 0
	}

	// The last piece might be shorter
	if index == t.NumPieces()-1 {
		totalLength := t.TotalLength()
		if totalLength%t.Info.PieceLength == 0 {
			return t.Info.PieceLength
		}
		return totalLength % t.Info.PieceLength
	}

	return t.Info.PieceLength
}

// TotalLength returns the total size of all files in the torrent
func (t *TorrentFile) TotalLength() int64 {
	if t.Info.Length > 0 {
		// Single file mode
		return t.Info.Length
	}

	// Multiple files mode
	var totalLength int64
	for _, file := range t.Info.Files {
		totalLength += file.Length
	}
	return totalLength
}
