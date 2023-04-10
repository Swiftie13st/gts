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
)

var Conf = new(AppConfig)

type AppConfig struct {
	Mode          string `mapstructure:"mode"`
	Name          string `mapstructure:"name"`
	Version       string `mapstructure:"version"`
	StartTime     string `mapstructure:"start_time"`
	Ip            string `mapstructure:"ip"`
	Port          int    `mapstructure:"port"`
	IpVersion     string `mapstructure:"ip_version"`
	MaxConn       int    `mapstructure:"max_conn"`
	MaxPacketSize uint32 `mapstructure:"max_packet_size"`
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
