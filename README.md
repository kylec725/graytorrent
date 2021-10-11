# graytorrent
BitTorrent engine implemented in [Go](https://golang.org)

## Features
- [BitTorrent Protocol](https://www.bittorrent.org/beps/bep_0003.html)
- [Multitracker Metadata Extension](https://www.bittorrent.org/beps/bep_0012.html)
- [UDP Trackers](https://www.bittorrent.org/beps/bep_0015.html)

## Installation
### Go
Make sure [Go is installed](https://golang.org/doc/install)
```
go install github.com/kylec725/graytorrent/cmd/gray`
```

Alternatively
```
git clone https://github.com/kylec725/graytorrent.git
cd graytorrent/cmd/gray
go install
```
Could also use `go build` or `go run main.go`

## Usage
Currently, graytorrent does not have a complete client. To use the torrenting functionality you will have to run graytorrent in single torrent download mode.
First, download the `.torrent` file for the torrent you want to use, then run `gray download pathtofile/examplefile.torrent` and graytorrent will start the torrent.

## Current Work
- Command line interface
- [Magnet Links](https://www.bittorrent.org/beps/bep_0009.html)
- Set-up graytorrent as a GRPC server
- Limit global number of connections

## Potential Features
- [Fast Extension](https://www.bittorrent.org/beps/bep_0006.html)
- [DHT](https://www.bittorrent.org/beps/bep_0005.html)
- [Peer Exchange (PEX)](https://www.bittorrent.org/beps/bep_0011.html)
- Protocol Encryption (MSE/PE)
- Rarest first requesting
- Use mmap for file operations

## Libraries Used
- [bencode-go](https://github.com/jackpal/bencode-go)
- [pkg/errors](https://github.com/pkg/errors)
- [logrus](https://github.com/sirupsen/logrus)
- [pflag](https://github.com/spf13/pflag)
- [cobra](https://github.com/spf13/cobra)
- [viper](https://github.com/spf13/viper)
- [testify](https://github.com/stretchr/testify)
- [logrus-prefixed-formatter](https://github.com/x-cray/logrus-prefixed-formatter)
