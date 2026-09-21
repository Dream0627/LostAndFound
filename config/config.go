// Package config 负责加载并监听项目配置（config/config.yaml）。
// 使用 Viper：它能把配置文件内容映射到结构体，并在文件变化时热更新。
// 结构体字段上的 mapstructure 标签，就是告诉 Viper 该字段对应 YAML 里的哪个键。
package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ServerConfig 对应 YAML 的 server 段：监听端口，以及对外可访问的基础 URL(用于拼接图片地址)。
type ServerConfig struct {
	Port          int    `mapstructure:"port"`
	PublicBaseURL string `mapstructure:"public_base_url"`
}

// DatabaseConfig 对应 database 段：MySQL 连接信息；Enabled 为 false 时程序将不连接数据库。
type DatabaseConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// Config 是配置总入口，把 server / database / jwt 三块聚合在一起。
type Config struct {
	Server    ServerConfig   `mapstructure:"server"`
	Database  DatabaseConfig `mapstructure:"database"`
	JWTConfig JWTConfig      `mapstructure:"jwt"`
}

// JWTConfig 对应 jwt 段：签发与校验令牌所用的密钥、有效期(秒)与签发者。
type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	ExpireSeconds int64  `mapstructure:"expires"`
	Issuer        string `mapstructure:"issuer"`
}

// Load 读取配置文件并解析成 Config。
// 流程：指定配置文件路径 -> 读入 -> 反序列化到结构体 -> 开启文件监听。
// WatchConfig 会在配置文件变化时重新反序列化；但 jwt 与 database 在启动时
// 已注入到各个 service，热更新不会让它们生效，必须重启服务（日志中会提示这一点）。
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
