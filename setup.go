package main

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/kylec725/graytorrent/connect"
	"github.com/kylec725/graytorrent/peer"
	"github.com/kylec725/graytorrent/peer/handshake"
	"github.com/kylec725/graytorrent/peer/message"
	"github.com/kylec725/graytorrent/torrent"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// ErrListener is used to close the peerListener safely
var ErrListener = errors.New("use of closed network connection")

func setupLog() {
	// Logging file
	logFile, err = os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	// logFile, err = os.OpenFile(filepath.Join(grayTorrentPath, "info.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Could not open log file")
	}

	// Set logging settings
	log.SetOutput(logFile)
	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	})
	if verbose {
		dualOutput := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(dualOutput)
	}
	log.SetLevel(logLevel)
}

func setupViper() {
	viper.SetDefault("torrent.path", ".")
	viper.SetDefault("torrent.autoseed", true)
	viper.SetDefault("network.portrange", [2]int{6881, 6889})
	viper.SetDefault("network.connections.globalMax", 300)
	viper.SetDefault("network.connections.torrentMax", 30)

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".") // Remove in the future
	viper.AddConfigPath(grayTorrentPath)
	viper.AddConfigPath("/etc/graytorrent")

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default config
			viper.SafeWriteConfig()
			log.Info("Config file written")
		} else {
			// Some other error was found
			log.Panic("Fatal error reading config file:", err)
		}
	}
}

// Binds a socket to some port for peers to contact us
func setupListen() {
	portRange := viper.GetIntSlice("network.portrange")
	port, err = connect.OpenPort(portRange)
	if err != nil {
		log.WithFields(log.Fields{
			"portrange": portRange,
			"error":     err.Error(),
		}).Warn("No open port found in port range, using random port")
	}

	service := ":" + strconv.Itoa(int(port))
	peerListener, err = net.Listen("tcp", service)
	if err != nil {
		panic("Could not bind to any port")
	}
	// Set global port
	port, err = connect.PortFromAddr(peerListener.Addr().String()) // Get actual port in case none in portrange were available
	if err != nil {
		panic("Could not find the binded port")
	}
}

func catchInterrupt(ctx context.Context, cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	select {
	case <-signalChan: // Cleanup on interrupt signal
		signal.Stop(signalChan)
		peerListener.Close()
		cancel()
		err = torrent.SaveAll(torrentList)
		if err != nil {
			log.WithField("error", err).Debug("Problem occurred while saving torrent management data")
		}
		log.Info("Graytorrent stopped")
		logFile.Close()
		os.Exit(1)
	case <-ctx.Done():
	}
}

// peerListen loops to listen for incoming connections of peers
func peerListen() {
	for {
		conn, err := peerListener.Accept()
		if err != nil { // Exit if the peerListener encounters an error
			if errors.Is(err, ErrListener) {
				log.WithField("error", err.Error()).Debug("Listener shutdown")
			}
			return
		}
		addr := conn.RemoteAddr().String()

		infoHash, err := handshake.Read(conn)
		if err != nil {
			log.WithFields(log.Fields{"peer": addr, "error": err.Error()}).Debug("Incoming peer handshake sequence failed")
			continue
		}

		// Check if the infohash matches any torrents we are serving
		for i := range torrentList {
			if bytes.Equal(infoHash[:], torrentList[i].Info.InfoHash[:]) {
				newPeer := peer.New(addr, conn, torrentList[i].Info)
				// Send back a handshake
				h := handshake.New(torrentList[i].Info)
				if _, err = newPeer.Conn.Write(h.Encode()); err != nil {
					log.WithFields(log.Fields{"peer": newPeer.String(), "error": err.Error()}).Debug("Incoming peer handshake sequence failed")
					break
				}

				// Send bitfield to the peer
				msg := message.Bitfield(torrentList[i].Info.Bitfield)
				if _, err = newPeer.Conn.Write(msg.Encode()); err != nil {
					log.WithFields(log.Fields{"peer": newPeer.String(), "error": err.Error()}).Debug("Sending bitfield failed")
					break
				}

				torrentList[i].IncomingPeers <- newPeer // Send to torrent session
				log.WithField("peer", newPeer.String()).Debug("Incoming peer was accepted")
			}
		}
	}
}
