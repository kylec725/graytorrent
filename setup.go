package main

import (
	"context"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/kylec725/graytorrent/connect"
	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	homeDir = os.Getenv("HOME")
)

func setupLog() {
	// Logging file
	logFile, err = os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	// logFile, err = os.OpenFile(homeDir + "/.config/graytorrent/info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Could not open log file info.log")
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
	viper.AddConfigPath(homeDir + "/.config/graytorrent")
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
	listener, err = net.Listen("tcp", service)
	if err != nil {
		panic("Could not bind to any port")
	}
	// Set global port
	port, err = connect.PortFromAddr(listener.Addr().String()) // Get actual port in case none in portrange were available
	if err != nil {
		panic("Could not find the binded port")
	}
}

func catchInterrupt(ctx context.Context, cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	select {
	case <-signalChan: // First signal, cleanup
		signal.Stop(signalChan)
		listener.Close()
		cancel()
		saveTorrents()
		log.Info("Graytorrent stopped")
		logFile.Close()
		os.Exit(1)
	case <-ctx.Done():
	}
	<-signalChan // Second signal, hard exit
	os.Exit(1)
}
