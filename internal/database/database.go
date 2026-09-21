// Package database 负责建立并维护 MySQL 连接。
// 本文件的关键点是“自动重连”：用后台协程每 5 秒 ping 一次数据库，
// 一旦发现断开就重新连接，从而在数据库短暂不可用时也能自愈，无需重启服务。
package database

import (
	"fmt"
	"sync"
	"time"

	"LAF/config" 
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// db 是全局数据库句柄；mu 是读写锁，用于在“重连时替换 db”与“业务并发读 db”之间保证并发安全。
// 使用 RWMutex 是因为绝大多数访问是并发读(RLock)，只有重连才写(Lock)，读多写少更高效。
var (
	db *gorm.DB
	mu sync.RWMutex 
)

// ConnectSQL 是对外的连接入口：先完成一次初始化连接，再启动后台自动重连协程。
func ConnectSQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	if err := initDB(cfg); err != nil {
		return nil, err
	}

	go autoReconnect(cfg)

	return db, nil
}

// 初始化连接
// initDB 真正建立(或重建)数据库连接。
// 使用写锁保护：因为它会替换全局 db，必须与并发的读取互斥。
func initDB(cfg config.DatabaseConfig) error {
	mu.Lock()
	defer mu.Unlock()

	if db != nil { // 已有旧连接时先关闭，避免连接泄漏
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", // 组装 MySQL DSN；parseTime=true 让时间字段能被正确扫描为 time.Time
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// 配置连接池
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}
// autoReconnect 是后台死循环：每 5 秒检查一次数据库连通性。
// 读取全局 db 时加读锁；一旦 ping 失败就调用 initDB 重连，实现故障自愈。
func autoReconnect(cfg config.DatabaseConfig) {
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
	defer ticker.Stop()

	for range ticker.C { // 无限循环：每收到一次 ticker 触发就做一轮健康检查
		mu.RLock()
		currentDB := db // 在读锁内取一份当前 db 引用，避免与“重连写 db”并发冲突
		mu.RUnlock()

		if currentDB != nil {
			sqlDB, err := currentDB.DB()
			if err == nil {
				if err := sqlDB.Ping(); err == nil { // ping 通说明连接健康，本轮无需处理
					continue
				}
			}
		}

		fmt.Println("\n数据库连接已断开,正在尝试重连...") // ping 失败说明连接已断开，准备重建连接
		if err := initDB(cfg); err != nil {
			fmt.Println("数据库重连失败:", err)
		} else {
			fmt.Println("数据库重连成功！")
		}
	}
}