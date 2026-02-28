// Package response 提供统一的 HTTP 响应格式
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// 响应状态码
const (
	CodeSuccess      = 0
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// SuccessPage 分页成功响应
func SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	httpStatus := http.StatusOK
	if code >= 400 && code < 500 {
		httpStatus = http.StatusBadRequest
	} else if code >= 500 {
		httpStatus = http.StatusInternalServerError
	}

	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	Error(c, CodeBadRequest, message)
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权"
	}
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: message,
	})
	c.Abort()
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "禁止访问"
	}
	c.JSON(http.StatusForbidden, Response{
		Code:    CodeForbidden,
		Message: message,
	})
	c.Abort()
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源不存在"
	}
	c.JSON(http.StatusNotFound, Response{
		Code:    CodeNotFound,
		Message: message,
	})
}

// ServerError 500 错误
func ServerError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeServerError,
		Message: message,
	})
}
