# graytorrent
BitTorrent engine implemented in [Go](https://golang.org)

## Features
- [BitTorrent Protocol](https://www.bittorrent.org/beps/bep_0003.html)
- [Multitracker Metadata Extension](https://www.bittorrent.org/beps/bep_0012.html)
- [UDP Trackers](https://www.bittorrent.org/beps/bep_0015.html)

## Installation
### Compile
Make sure [Go is installed](https://golang.org/doc/install)
```
git clone https://github.com/kylec725/graytorrent.git
cd graytorrent
go install
```
Could alternatively use `go build` or `go run main.go`

## Usage
Currently, graytorrent does not have a complete client. To use the torrenting functionality you will have to run graytorrent in single torrent download mode.
First, download the `.torrent` file for the torrent you want to use, then run `graytorrent -f pathtofile/examplefile.torrent` and graytorrent will start the torrent.

## Current Work
- Set-up graytorrent as a GRPC server
- Command line interface
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
- [viper](https://github.com/spf13/viper)
- [testify](https://github.com/stretchr/testify)
- [logrus-prefixed-formatter](https://github.com/x-cray/logrus-prefixed-formatter)
