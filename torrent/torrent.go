package torrent

import (
	"crypto/sha1"
	"errors"
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
