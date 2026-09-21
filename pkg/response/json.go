// Package response 定义全项目统一的 HTTP 响应“信封”。
// 设计目的：让所有接口无论成功还是失败，都返回同一种 JSON 结构，
// 前端只需要一套解析逻辑即可同时处理成功与失败，无需为每个接口单独适配。
// 信封包含三个字段：code(业务码，0 表示成功)、msg(提示语)、data(业务数据)。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope 是统一响应体结构。
// Data 使用 interface{}（空接口）作为类型，是因为不同接口返回的业务数据形态各异
// （对象、数组、字符串等），空接口可以容纳任意类型；
// JSON 标签决定序列化到响应体时字段名分别是 code / msg / data。
type Envelope struct {
	Code    int         `json:"code"`
	Msg 	string      `json:"msg"`
	Data    interface{} `json:"data"`
}

// JSON 把信封按指定的 HTTP 状态码 + 业务码写回响应。
// 注意区分两个“码”：httpStatus 是 HTTP 层状态码（如 200/400），
// code 是响应体内部的业务码；本项目习惯让二者保持一致（见下面的 Error）。
func JSON(c *gin.Context, httpStatus, code int, msg string, data interface{}) {
	c.JSON(httpStatus, Envelope{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}
// Success 是成功响应的快捷方法：HTTP 状态 200、业务码固定为 0、msg 固定为 "success"。
// 所有 handler 处理成功后都调用它，保证成功响应格式完全一致。
func Success(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, 0, "success", data)
}

// Error 是失败响应的快捷方法。
// 它把业务码直接当作 HTTP 状态码使用；若业务码不是合法 HTTP 状态码（100~599），
// 则回退为 500，避免产生浏览器/网关无法识别的状态码。
func Error(c *gin.Context, code int, msg string) {
	httpStatus := code
	if httpStatus < 100 || httpStatus > 599 {
		httpStatus = http.StatusInternalServerError
	}
	JSON(c, httpStatus, code, msg, nil)
}