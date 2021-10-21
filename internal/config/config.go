package config

import (
	"github.com/kylec725/graytorrent/internal/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config represents the configuration of the program
type Config struct {
	Torrent TorrentConfig
	Network NetworkConfig
}

// TorrentConfig provides settings for torrents
type TorrentConfig struct {
	DefaultPath string `mapstructure:"default_path"`
	AutoSeed    bool   `mapstructure:"auto_seed"`
}

// NetworkConfig provides settings for network options
type NetworkConfig struct {
	ListenerPort          []int `mapstructure:"listener_port"`
	ServerPort            int   `mapstructure:"server_port"`
	MaxGlobalConnections  int   `mapstructure:"max_global_connections"`
	MaxTorrentConnections int   `mapstructure:"max_torrent_connections"`
}

// InitConfig initializes the config file and default values
func InitConfig() {
	viper.SetDefault("torrent.default_path", ".")
	viper.SetDefault("torrent.auto_seed", true)
	viper.SetDefault("network.listener_port", [2]int{6881, 6889})
	viper.SetDefault("network.max_global_connections", 300)
	viper.SetDefault("network.max_torrent_connections", 30)
	viper.SetDefault("network.server_port", 7001)

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
			log.Fatal("Fatal error reading config file:", err)
		}
	}
}

// GetConfig returns an initialized Config struct
func GetConfig() (cfg Config) {
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatal("Error parsing config:", err)
	}
	return
}
