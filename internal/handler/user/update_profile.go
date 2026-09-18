package user

import (
	"LAF/internal/service"
	"LAF/pkg/response"
	//"errors"
	//"strconv"

	//"fmt"

	//"strings"
	//"time"
	//"LAF/internal/model"
	//"LAF/internal/repository"
	"LAF/internal/middleware"
	"LAF/pkg/apperror"

	"github.com/gin-gonic/gin"
)

func UpdateProfile(userService *service.UserService) gin.HandlerFunc { 
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)
		
		var input service.UpdateProfileInput
		if err := c.ShouldBindJSON(&input); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}
		
		if err := userService.UpdateProfile(nowUserID, input); err != nil {
			apperror.AbortWithException(c, apperror.ServerError, err)
			return
		}

		user, err := userService.GetProfile(nowUserID)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, user)
	}
	
}