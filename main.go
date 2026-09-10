package main

import (
	"fmt"
	"os"

	"LAF/config"
	"LAF/internal/database"
	"LAF/internal/router"

	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("load config failed:", err)
		os.Exit(1)
	}

	var db *gorm.DB

	if cfg.Database.Enabled {
		db, err = database.ConnectSQL(cfg.Database)
		if err != nil {
			fmt.Fprintln(os.Stderr, "connect database failed:", err)
			os.Exit(1)
		}

		sqlDB, err := db.DB()
		if err != nil {
			fmt.Fprintln(os.Stderr, "get database connection failed:", err)
			os.Exit(1)
		}
		defer sqlDB.Close()
	}

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	engine := router.New(db, cfg.JWTConfig, cfg.Server.PublicBaseURL)
	if err := engine.Run(port); err != nil {
		fmt.Println("start server failed:", err)
	}
}
