package user

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(userService *service.UserService) gin.HandlerFunc { 
	return func(c *gin.Context) {
		var request LoginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		result, err := userService.Login(request.Username, request.Password)
		if err != nil {
			apperror.AbortWithError(c, err)
			return 
		}

		response.Success(c, result)
	}
}