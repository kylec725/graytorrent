package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/kylec725/graytorrent/internal/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print events to stdout")
	cobra.OnInitialize(initDirs, initConfig)
}

var (
	logLevel = log.InfoLevel
	logFile  *os.File

	// Flags
	debug      bool
	verbose    bool
	magnetLink string
	isInfoHash bool

	rootCmd = &cobra.Command{
		Use:     "gray",
		Short:   "graytorrent is a BitTorrent engine",
		Long:    `graytorrent is an engine that implements the BitTorrent Protocol and allows for the management of torrents.`,
		Version: "0.20",
	}
)

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// initDirs initializes any necessary management directories
func initDirs() {
	err := os.MkdirAll(common.SavePath, os.ModePerm) // SavePath includes GrayPath
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Could not create necessary directories")
	}
}

func initLog() {
	// Logging file
	logFile, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	// logFile, err = os.OpenFile(filepath.Join(common.GrayTorrentPath, "info.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Could not open log file")
	}

	// Set logging settings
	log.SetOutput(logFile)
	log.SetFormatter(&prefixed.TextFormatter{
		// TimestampFormat: "2006-01-02 15:04:05",
		TimestampFormat: "15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	})
	// Set flags
	if verbose {
		dualOutput := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(dualOutput)
	}
	if debug {
		logLevel = log.TraceLevel
	}
	log.SetLevel(logLevel)
}

func initConfig() {
	viper.SetDefault("torrent.path", ".")
	viper.SetDefault("torrent.autoseed", true)
	viper.SetDefault("network.portrange", [2]int{6881, 6889})
	viper.SetDefault("network.connections.globalMax", 300)
	viper.SetDefault("network.connections.torrentMax", 30)
	viper.SetDefault("server.port", 7001)

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(common.GrayTorrentPath)
	// viper.AddConfigPath("/etc/gray")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default config
			viper.SafeWriteConfig()
			log.Info("Config file written at " + common.GrayTorrentPath)
		} else {
			// Some other error was found
			log.Panic("Fatal error reading config file:", err)
		}
	}
}
