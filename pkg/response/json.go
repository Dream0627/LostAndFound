package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code    int         `json:"code"`
	Msg 	string      `json:"msg"`
	Data    interface{} `json:"data"`
}

func JSON(c *gin.Context, httpStatus, code int, msg string, data interface{}) {
	c.JSON(httpStatus, Envelope{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}
func Success(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, 0, "success", data)
}

func Error(c *gin.Context, code int, msg string) {
	JSON(c, code, code, msg, nil)
}