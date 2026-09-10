package database

import (
	"fmt"
	"sync"
	"time"

	"LAF/config" 
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
	mu sync.RWMutex 
)

func ConnectSQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	if err := initDB(cfg); err != nil {
		return nil, err
	}

	go autoReconnect(cfg)

	return db, nil
}

// 初始化连接
func initDB(cfg config.DatabaseConfig) error {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
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
func autoReconnect(cfg config.DatabaseConfig) {
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
	defer ticker.Stop()

	for range ticker.C {
		mu.RLock()
		currentDB := db
		mu.RUnlock()

		if currentDB != nil {
			sqlDB, err := currentDB.DB()
			if err == nil {
				if err := sqlDB.Ping(); err == nil {
					continue
				}
			}
		}

		fmt.Println("\n数据库连接已断开,正在尝试重连...")
		if err := initDB(cfg); err != nil {
			fmt.Println("数据库重连失败:", err)
		} else {
			fmt.Println("数据库重连成功！")
		}
	}
}