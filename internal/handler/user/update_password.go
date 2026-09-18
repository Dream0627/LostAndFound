package user

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

type UpdatePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func UpdatePassword(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)

		var request UpdatePasswordRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		input := service.UpdatePasswordInput{
			OldPassword:     request.OldPassword,
			NewPassword:     request.NewPassword,
			ConfirmPassword: request.ConfirmPassword,
		}
		if err := userService.UpdatePassword(nowUserID, input); err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, nil)
	}
}
