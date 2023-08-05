package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kylec725/graytorrent/internal/common"
	"github.com/kylec725/graytorrent/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func init() {
	cobra.OnInitialize(initDirs, config.InitConfig)
}

var (
	version string

	logLevel = log.InfoLevel
	logFile  *os.File

	// Flags
	debug      bool
	magnet     bool
	isInfoHash bool
	directory  string
	rmFiles    bool

	rootCmd = &cobra.Command{
		Use:     "gray",
		Short:   "graytorrent is a BitTorrent engine",
		Long:    `graytorrent is an engine that implements the BitTorrent Protocol and allows for the management of torrents.`,
		Version: version,
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
	// logFile, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logFile, err := os.OpenFile(filepath.Join(common.GrayTorrentPath, "info.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
	if debug {
		logLevel = log.TraceLevel
	}
	log.SetLevel(logLevel)
}
