# graytorrent
BitTorrent engine implemented in [Go](https://golang.org)

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

## In Progress
- Setting up graytorrent to act as a GRPC server
- Limit global number of connections

## Future Goals
- Add Protocol Encryption (MSE/PE)
- Extensions (BEP6, DHT/PEX)
- Efficient rarest first requesting algorithm
- Switch to mmap file operations (possibly directly writing blocks?)

## Dependencies
- [bencode-go](https://github.com/jackpal/bencode-go)
- [pkg/errors](https://github.com/pkg/errors)
- [logrus](https://github.com/sirupsen/logrus)
- [pflag](https://github.com/spf13/pflag)
- [viper](https://github.com/spf13/viper)
- [testify](https://github.com/stretchr/testify)
- [logrus-prefixed-formatter](https://github.com/x-cray/logrus-prefixed-formatter)
