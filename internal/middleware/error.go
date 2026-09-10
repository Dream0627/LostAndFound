package middleware

import (
	"github.com/gin-gonic/gin"

	"LAF/pkg/apperror"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		apperror.Handle(c, c.Errors.Last().Err)
	}
}