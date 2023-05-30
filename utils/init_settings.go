/**
  @author: Bruce
  @since: 2023/3/17
  @desc: //setting
**/

package utils

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"time"
)

var Conf = new(AppConfig)

type AppConfig struct {
	Mode              string `mapstructure:"mode"`
	Name              string `mapstructure:"name"`
	Version           string `mapstructure:"version"`
	StartTime         string `mapstructure:"start_time"`
	TCPMode           bool   `mapstructure:"tcp_mode"`
	WSMode            bool   `mapstructure:"ws_mode"`
	Ip                string `mapstructure:"ip"`
	Port              int    `mapstructure:"port"`
	WsPort            int    `mapstructure:"ws_port"`
	IpVersion         string `mapstructure:"ip_version"`
	MaxConn           int    `mapstructure:"max_conn"`
	MaxPacketSize     uint32 `mapstructure:"max_packet_size"`
	WorkerPoolSize    int    `mapstructure:"worker_pool_size"`
	MaxWorkerTaskLen  uint64 `mapstructure:"max_worker_task_len"`
	HeartbeatMaxTime  int    `mapstructure:"heartbeat_max_time"`
	HeartbeatInterval int    `mapstructure:"heartbeat_interval"`
	WorkerId          int64  `mapstructure:"worker_id"`
	DatacenterId      int64  `mapstructure:"datacenter_id"`
	LogLevel          string `mapstructure:"log_level"`
	LogPath           string `mapstructure:"log_path"`
}

func (g *AppConfig) GetHeartbeatInterval() time.Duration {
	return time.Duration(g.HeartbeatInterval) * time.Second
}
func (g *AppConfig) GetHeartbeatMaxTime() time.Duration {
	return time.Duration(g.HeartbeatMaxTime) * time.Second
}

func InitSettings(path string) {
	viper.SetConfigFile(path)

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("###配置文件修改###")
		err := viper.Unmarshal(&Conf)
		if err != nil {
			return
		}
	})

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("ReadInConfig failed, err: %v", err))
	}
	if err := viper.Unmarshal(&Conf); err != nil {
		panic(fmt.Errorf("unmarshal to Conf failed, err:%v", err))
	}
}
