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

func New(db *gorm.DB, jwtConfig config.JWTConfig, publicBaseURL string) *gin.Engine {
	engine := gin.Default()
	engine.Use(middleware.ErrorHandler())
	engine.Static("/uploads", "./uploads")

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository, jwtConfig)
	postRepository := repository.NewPostRepository(db)
	postService := service.NewPostService(postRepository)
	postAdminService := service.NewPostAdminService(postRepository)
	commentRepository := repository.NewCommentRepository(db)
	commentService := service.NewCommentService(commentRepository, postRepository)
	//postAdminRepository := repository.NewPostAdminRepository(db)
	//postAdminService := service.NewPostAdminService(postAdminRepository)
	//mainAdminRepository := repository.NewMainAdminRepository(db)
	//mainAdminService := service.NewMainAdminService(mainAdminRepository)

	auth := engine.Group("/api/v1/auth")
	auth.POST("/register", userhandler.Register(userService))
	auth.POST("/login", userhandler.Login(userService))
	auth.GET("/profile", middleware.Auth(jwtConfig), userhandler.GetProfile(userService))
	auth.PATCH("/profile", middleware.Auth(jwtConfig), userhandler.UpdateProfile(userService))
	auth.PATCH("/password", middleware.Auth(jwtConfig), userhandler.UpdatePassword(userService))

	post := engine.Group("/api/v1/posts")
	post.GET("", middleware.OptionalAuth(jwtConfig), posthandler.ListPosts(postService))
	post.POST("", middleware.Auth(jwtConfig), posthandler.Create(postService, publicBaseURL))
	post.GET("/:post_id", middleware.OptionalAuth(jwtConfig), posthandler.GetPost(postService))
	post.DELETE("/:post_id", middleware.Auth(jwtConfig),posthandler.DeletePost(postService))
	post.PATCH("/:post_id/recover", middleware.Auth(jwtConfig),posthandler.RecoverPost(postService))
	post.PATCH("/:post_id/review", middleware.Auth(jwtConfig), middleware.RequireRole([]string{"postadmin", "mainadmin"}), postadminhandler.ReviewPost(postAdminService))
	post.GET("/:post_id/comments", commenthandler.List(commentService))

	comment := engine.Group("/api/v1/comments")
	comment.POST("", middleware.Auth(jwtConfig), commenthandler.Create(commentService))
	comment.DELETE("/:comment_id", middleware.Auth(jwtConfig), commenthandler.Delete(commentService))

	admin := engine.Group("/api/v1/admin")
	admin.PATCH("/posts/:post_id/status", middleware.Auth(jwtConfig), middleware.RequireRole([]string{"postadmin", "mainadmin"}), postadminhandler.UpdatePostStatus(postAdminService))
	admin.GET("/posts/deleted", middleware.Auth(jwtConfig), middleware.RequireRole([]string{"postadmin", "mainadmin"}), postadminhandler.ListDeletedPosts(postAdminService))

	// postadmin := admin.Group("/postadmin")

	// mainadmin := admin.Group("/mainadmin")

	return engine
}
