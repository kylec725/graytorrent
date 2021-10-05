# graytorrent
BitTorrent client implemented in golang

## Installation
### Manual
First make sure [Go is installed](https://golang.org/doc/install)
```
git clone https://github.com/kylec725/graytorrent.git
cd graytorrent
go install
```
Could alternatively use `go build` or `go run main.go`

## In Progress
- Setting up graytorrent to act as a GRPC server
- Optimistic Unchoking
- Limit number of connections

## Future Goals
- Extensions (BEP6, DHT/PEX)
- Efficient rarest first requesting algorithm
- Fine-tune the adaptive download rate

## Dependencies
- [bencode-go](https://github.com/jackpal/bencode-go)
- [pkg/errors](https://github.com/pkg/errors)
- [logrus](https://github.com/sirupsen/logrus)
- [pflag](https://github.com/spf13/pflag)
- [viper](https://github.com/spf13/viper)
- [testify](https://github.com/stretchr/testify)
- [logrus-prefixed-formatter](https://github.com/x-cray/logrus-prefixed-formatter)
