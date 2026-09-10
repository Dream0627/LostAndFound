package apperror

import (
	"errors"

	"github.com/gin-gonic/gin"

	"LAF/pkg/response"
)

const InternalErrorKey = "internal_error"

func AbortWithException(c *gin.Context, apiError *Error, err error) {
	if err != nil {
		c.Set(InternalErrorKey, err)
	}
	_ = c.Error(apiError)
	c.Abort()
}

func Handle(c *gin.Context, err error) {
	apiError := ServerError
	var target *Error
	if errors.As(err, &target) {
		apiError = target
	}
	response.Error(c, apiError.Code, apiError.Msg)
}

func AbortWithError(c *gin.Context, apiError error) {
	c.Set(InternalErrorKey, apiError)
	_ = c.Error(apiError)
	c.Abort()
}