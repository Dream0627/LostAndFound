package router

import (
	"LAF/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	posthandler "LAF/internal/handler/post"
	userhandler "LAF/internal/handler/user"

	//postadminhandler "LAF/internal/handler/postadmin"
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
	//postAdminRepository := repository.NewPostAdminRepository(db)
	//postAdminService := service.NewPostAdminService(postAdminRepository)
	//mainAdminRepository := repository.NewMainAdminRepository(db)
	//mainAdminService := service.NewMainAdminService(mainAdminRepository)

	auth := engine.Group("/api/v1/auth")
	auth.POST("/register", userhandler.Register(userService))
	auth.POST("/login", userhandler.Login(userService))

	post := engine.Group("/api/v1/posts")
	post.POST("", middleware.Auth(jwtConfig), posthandler.Create(postService, publicBaseURL))
	post.DELETE("/:post_id", middleware.Auth(jwtConfig),posthandler.DeletePost(postService))

	// admin := engine.Group("/api/v1/admin")
	// admin.Use(middleware.Auth(jwtConfig), middleware.RequireRole([]string{"postadmin", "mainadmin"}))
	// admin.DELETE("/posts/:post_id", postadminhandler.DeletePost(postAdminService))

	// postadmin := admin.Group("/postadmin")
	// postadmin.Use(middleware.Auth(jwtConfig), middleware.RequireRole([]string{"postadmin"}))

	// mainadmin := admin.Group("/mainadmin")
	// mainadmin.Use(middleware.Auth(jwtConfig), middleware.RequireRole([]string{"mainadmin"}))

	return engine
}
