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

func GetProfile(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)
		
		user, err := userService.GetProfile(nowUserID)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, user)
	}
}