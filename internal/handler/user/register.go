// 本文件对应“注册”接口。
package user

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// RegisterRequest 是注册请求体。
// 每个字段同时带 form 与 json 标签：允许用表单或 JSON 提交，灵活性更高。
// binding:"required" 表示这些字段都必传。
type RegisterRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Name	 string `form:"name" json:"name" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
	Role     string `form:"role" json:"role" binding:"required"`
}

// Register 是“注册”的处理器工厂。
func Register(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request RegisterRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		createdUser, err := userService.Register(service.RegisterInput{ // 把请求体转成业务入参并调用业务层
			Username: request.Username,
			Name:     request.Name,
			Password: request.Password,
			Role:     request.Role,
		})
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdUser) // 返回新建的用户(不含密码哈希)
	}
}