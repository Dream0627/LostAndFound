package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port          int    `mapstructure:"port"`
	PublicBaseURL string `mapstructure:"public_base_url"`
}

type DatabaseConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type Config struct {
	Server    ServerConfig   `mapstructure:"server"`
	Database  DatabaseConfig `mapstructure:"database"`
	JWTConfig JWTConfig      `mapstructure:"jwt"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	ExpireSeconds int64  `mapstructure:"expires"`
	Issuer        string `mapstructure:"issuer"`
}

func Load() (*Config, error) {
	viper.SetConfigFile("config/config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("读取配置失败", err)
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println("解析配置失败", err)
		return nil, err
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("配置已改变", e.Name)
		if err := viper.Unmarshal(&cfg); err != nil {
			fmt.Println("重新加载配置失败", err)
			return
		}
		fmt.Println("重新加载配置成功", e.Name)
		fmt.Println("注意：jwt/database 配置已在启动时注入各服务，本次热更新不会生效，需重启服务")
	})

	return &cfg, nil
}
