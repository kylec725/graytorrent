# graytorrent
BitTorrent client implemented in golang

## In Progress...
- Fine-tune the adaptive download rate
- Restructure the peers, trackers, and torrents as formal finite-state machines
- UDP Tracker support
- Tracker scraping

## Potential Goals
- Extensions (BEP6, DHT/PEX)
- Efficient rarest first requesting algorithm
- Use Google Protocol Buffers for network interactions

## Dependencies
- [bencode-go](github.com/jackpal/bencode-go)
- [pkg/errors](github.com/pkg/errors)
- [logrus](github.com/sirupsen/logrus)
- [pflag](github.com/spf13/pflag)
- [viper](github.com/spf13/viper)
- [testify](github.com/stretchr/testify)
- [logrus-prefixed-formatter](github.com/x-cray/logrus-prefixed-formatter)
