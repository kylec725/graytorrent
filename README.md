# graytorrent
BitTorrent engine implemented in [Go](https://golang.org)

## Features
- [BitTorrent Protocol](https://www.bittorrent.org/beps/bep_0003.html)
- Command line interface
- [Multitracker Metadata Extension](https://www.bittorrent.org/beps/bep_0012.html)
- [UDP Trackers](https://www.bittorrent.org/beps/bep_0015.html)

## Installation
### Go
Make sure [Go is installed](https://golang.org/doc/install)
```
go install github.com/kylec725/graytorrent/cmd/gray
```

Alternatively
```
git clone https://github.com/kylec725/graytorrent.git
cd graytorrent/cmd/gray
go install
```
Could also use `go build` or `go run main.go`

## Usage
graytorrent has a command line interface. You can see the available commands by entering `gray`.

### Starting gray
To start the gray torrent server
```
gray server start
```
Stop the server with
```
gray server stop
```

### Adding torrents
Currently, graytorrent only handles `.torrent` files. First, download the `.torrent` file for the torrent you want to use,
then add it to graytorrent's list of managed torrents with
```
gray add filepath/example.torrent
```

### List managed torrents
You can view all managed torrents with
```
gray ls
```

### Controlling the managed torrents
To start or stop the upload/download of the torrents you can use the number IDs that are listed or their infohash (with the `-i` flag) to select them.
```
gray start ID
gray stop ID
```

### Remove a manged torrent with
```
gray rm ID
```

## Current Work
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
