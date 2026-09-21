// Package router 是项目的“唯一装配点”与路由注册中心。
// 它按 仓库(repository) -> 服务(service) -> 处理器(handler) 的顺序自下而上地把依赖 new 出来，
// 再把每个 handler 绑定到对应的 HTTP 方法与路径上。
// 这样做的好处：依赖关系集中可见，各层不需要自己去找依赖，也便于替换实现或做测试。
package router

import (
	"LAF/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	posthandler "LAF/internal/handler/post"
	userhandler "LAF/internal/handler/user"
	commenthandler "LAF/internal/handler/comment"

	postadminhandler "LAF/internal/handler/postadmin"
	//mainadminhandler "LAF/internal/handler/mainadmin"
	"LAF/internal/middleware"
	"LAF/internal/repository"
	"LAF/internal/service"
)

// New 构建并返回配置好的 gin.Engine。参数：数据库句柄、JWT 配置、图片对外基础地址。
func New(db *gorm.DB, jwtConfig config.JWTConfig, publicBaseURL string) *gin.Engine {
	engine := gin.Default()
	engine.Use(middleware.ErrorHandler()) // 注册全局错误处理中间件，统一兜底错误响应
	engine.Static("/uploads", "./uploads") // 把本地上传目录映射成静态资源，使图片可通过 /uploads/... 直接访问

	userRepository := repository.NewUserRepository(db) // 装配阶段：先建仓库
	userService := service.NewUserService(userRepository, jwtConfig) // 再建服务，注入仓库
	postRepository := repository.NewPostRepository(db)
	postService := service.NewPostService(postRepository)
	postAdminService := service.NewPostAdminService(postRepository)
	commentRepository := repository.NewCommentRepository(db)
	commentService := service.NewCommentService(commentRepository, postRepository)
	//postAdminRepository := repository.NewPostAdminRepository(db)
	//postAdminService := service.NewPostAdminService(postAdminRepository)
	//mainAdminRepository := repository.NewMainAdminRepository(db)
	//mainAdminService := service.NewMainAdminService(mainAdminRepository)

	auth := engine.Group("/api/v1/auth") // 认证相关路由分组(注册/登录/资料/改密)
	auth.POST("/register", userhandler.Register(userService))
	auth.POST("/login", userhandler.Login(userService))
	auth.GET("/profile", middleware.Auth(jwtConfig), userhandler.GetProfile(userService))
	auth.PATCH("/profile", middleware.Auth(jwtConfig), userhandler.UpdateProfile(userService))
	auth.PATCH("/password", middleware.Auth(jwtConfig), userhandler.UpdatePassword(userService))

	post := engine.Group("/api/v1/posts") // 帖子相关路由分组
	post.GET("", middleware.OptionalAuth(jwtConfig), posthandler.ListPosts(postService))
	post.POST("", middleware.Auth(jwtConfig), posthandler.Create(postService, publicBaseURL))
	post.GET("/:post_id", middleware.OptionalAuth(jwtConfig), posthandler.GetPost(postService))
	post.DELETE("/:post_id", middleware.Auth(jwtConfig),posthandler.DeletePost(postService))
	post.PATCH("/:post_id/recover", middleware.Auth(jwtConfig),posthandler.RecoverPost(postService))
	post.PATCH("/:post_id/review", middleware.Auth(jwtConfig), middleware.RequireRole([]string{"postadmin", "mainadmin"}), postadminhandler.ReviewPost(postAdminService)) // 审核帖子：先登录校验，再要求管理员角色
	post.GET("/:post_id/comments", commenthandler.List(commentService)) // 评论列表仍挂在帖子下(读操作，语义上属于某帖的评论)

	comment := engine.Group("/api/v1/comments") // 评论相关路由分组
	comment.POST("", middleware.Auth(jwtConfig), commenthandler.Create(commentService)) // 发表评论已迁到评论分组(写操作归属评论模块)
	comment.DELETE("/:comment_id", middleware.Auth(jwtConfig), commenthandler.Delete(commentService)) // 删除评论

	admin := engine.Group("/api/v1/admin") // 管理员路由分组(叠加角色校验)
	admin.PATCH("/posts/:post_id/status", middleware.Auth(jwtConfig), middleware.RequireRole([]string{"postadmin", "mainadmin"}), postadminhandler.UpdatePostStatus(postAdminService))
	admin.GET("/posts/deleted", middleware.Auth(jwtConfig), middleware.RequireRole([]string{"postadmin", "mainadmin"}), postadminhandler.ListDeletedPosts(postAdminService))

	// postadmin := admin.Group("/postadmin")

	// mainadmin := admin.Group("/mainadmin")

	return engine
}
