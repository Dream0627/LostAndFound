// 程序入口。整体顺序：加载配置 -> 连接数据库 -> 构建路由与依赖 -> 启动 HTTP 服务。
// 任何一个前置步骤失败都会打印错误并退出，避免服务“带病启动”。
package main

import (
	"fmt"
	"os"

	"LAF/config"
	"LAF/internal/database"
	"LAF/internal/router"

	"gorm.io/gorm"
)

// main 串联整个启动流程。
func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("load config failed:", err)
		os.Exit(1)
	}

	var db *gorm.DB

	if cfg.Database.Enabled { // 仅当配置里开启数据库时才建立连接
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
		defer sqlDB.Close() // 进程退出时关闭底层数据库连接池，释放资源
	}

	port := fmt.Sprintf(":%d", cfg.Server.Port)
	engine := router.New(db, cfg.JWTConfig, cfg.Server.PublicBaseURL) // 组装“仓库 -> 服务 -> 处理器”并注册所有路由
	if err := engine.Run(port); err != nil {
		fmt.Println("start server failed:", err)
	}
}
