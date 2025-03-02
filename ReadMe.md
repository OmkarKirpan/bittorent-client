# Go BitTorrent Client

## Overview

This repository contains a BitTorrent client written in Go, developed as a learning project to understand the BitTorrent protocol, file sharing mechanisms, and peer-to-peer networking.

## Features

- Decode `.torrent` files
- Discover and connect to peers
- Perform the BitTorrent handshake
- Download and verify file pieces
- Support magnet links and metadata exchange
- Extend protocol with additional features

## Roadmap

The project is structured into incremental steps:

1. **Introduction** - Overview of the project and its objectives.
2. **Repository Setup** - Setting up the Go environment and dependencies.
3. **Decode bencoded strings** - Implement parsing of bencoded strings.
4. **Decode bencoded integers** - Implement parsing of bencoded integers.
5. **Decode bencoded lists** - Implement parsing of bencoded lists.
6. **Decode bencoded dictionaries** - Implement parsing of bencoded dictionaries.
7. **Parse torrent file** - Extract metadata from `.torrent` files.
8. **Calculate info hash** - Compute the SHA-1 hash of the info dictionary.
9. **Piece hashes** - Validate file integrity using piece hashes.
10. **Discover peers** - Locate peers using tracker and DHT.
11. **Peer handshake** - Establish connections with peers.
12. **Download a piece** - Implement piece request and transfer.
13. **Download the whole file** - Assemble downloaded pieces into a complete file.
14. **Parse magnet link** - Extract metadata from magnet links.
15. **Announce extension support** - Advertise extended protocol support.
16. **Send extension handshake** - Negotiate additional features with peers.
17. **Receive extension handshake** - Handle incoming extension handshakes.
18. **Request metadata** - Fetch metadata from peers when using magnet links.
19. **Receive metadata** - Parse and store metadata received from peers.
20. **Download a piece** - Fetch individual file pieces via peer connections.
21. **Download the whole file** - Assemble the full file from downloaded pieces.

## Prerequisites

- Go 1.22+
- Basic understanding of networking and file I/O

## Getting Started

1. Clone the repository:

   ```sh
   git clone https://github.com/yourusername/go-bittorrent.git
   cd go-bittorrent
   ```

2. Install dependencies:

   ```sh
   go mod tidy
   ```

3. Run the project:

   ```sh
   go run main.go
   ```

## Contributing

Contributions are welcome! Feel free to open issues and submit pull requests.

## License

This project is licensed under the MIT License.

## Author

Name: Omkar Kirpan  
Website: [https://omkarkirpan.com](https://omkarkirpan.com)
