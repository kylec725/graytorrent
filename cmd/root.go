package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print events to stdout")
	cobra.OnInitialize(initLog, initConfig, viper.WatchConfig)
}

var (
	logLevel        = log.TraceLevel // InfoLevel || DebugLevel || TraceLevel
	logFile         *os.File
	grayTorrentPath = filepath.Join(os.Getenv("HOME"), ".config", "graytorrent")

	// Flags
	verbose bool

	rootCmd = &cobra.Command{
		Use:   "graytorrent",
		Short: "graytorrent is a BitTorrent engine",
		Long:  `An engine that implements the BitTorrent Protocol and allows for the management of torrents.`,
		// Run: func(cmd *cobra.Command, args []string) {
		// 	// Do Stuff Here
		// 	fmt.Println("hello root command")
		// },
	}
)

// Execute runs the root command
func Execute() {
	initDirs()
	// log.Info("Graytorrent started")
	defer func() {
		// log.Info("Graytorrent stopped")
		logFile.Close()
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// initDirs initializes any necessary management directories
func initDirs() {
	err := os.MkdirAll(grayTorrentPath, os.ModePerm)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Could not create necessary directories")
	}
}

func initLog() {
	// Logging file
	logFile, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	// logFile, err = os.OpenFile(filepath.Join(grayTorrentPath, "info.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
	if verbose {
		dualOutput := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(dualOutput)
	}
	log.SetLevel(logLevel)
}

func initConfig() {
	viper.SetDefault("torrent.path", ".")
	viper.SetDefault("torrent.autoseed", true)
	viper.SetDefault("network.portrange", [2]int{6881, 6889})
	viper.SetDefault("network.connections.globalMax", 300)
	viper.SetDefault("network.connections.torrentMax", 30)

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	// viper.AddConfigPath(".") // Remove in the future
	viper.AddConfigPath(grayTorrentPath)
	viper.AddConfigPath("/etc/graytorrent")

	if err := viper.ReadInConfig(); err != nil {
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
