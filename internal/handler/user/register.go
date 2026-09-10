package user

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

type RegisterRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Name	 string `form:"name" json:"name" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Role     string `form:"role" json:"role" binding:"required"`
}

func Register(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request RegisterRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		createdUser, err := userService.Register(service.RegisterInput{
			Username: request.Username,
			Name:     request.Name,
			Password: request.Password,
			Role:     request.Role,
		})
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdUser)
	}
}