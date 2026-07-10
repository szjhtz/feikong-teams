package handler

import (
	"fkteams/internal/domain/apperror"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一 API 响应结构
type Response struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
}

// OK 返回成功响应
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

// Created 返回资源创建成功响应。
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "success", Data: data})
}

// Fail 返回失败响应
func Fail(c *gin.Context, httpCode int, msg string) {
	c.JSON(httpCode, Response{Code: 1, ErrorCode: errorCodeForHTTPStatus(httpCode), Message: msg})
}

func errorCodeForHTTPStatus(status int) string {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return string(apperror.CodeInvalidArgument)
	case http.StatusUnauthorized:
		return string(apperror.CodeUnauthorized)
	case http.StatusForbidden:
		return string(apperror.CodeForbidden)
	case http.StatusNotFound, http.StatusGone:
		return string(apperror.CodeNotFound)
	case http.StatusConflict:
		return string(apperror.CodeConflict)
	case http.StatusRequestEntityTooLarge, http.StatusTooManyRequests:
		return string(apperror.CodeResourceLimit)
	case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		return string(apperror.CodeUnavailable)
	default:
		return string(apperror.CodeInternal)
	}
}

// FailError 根据稳定应用错误码映射 HTTP 状态和错误响应。
func FailError(c *gin.Context, err error) {
	code := apperror.CodeOf(err)
	status := http.StatusInternalServerError
	switch code {
	case apperror.CodeInvalidArgument:
		status = http.StatusBadRequest
	case apperror.CodeNotFound:
		status = http.StatusNotFound
	case apperror.CodeConflict:
		status = http.StatusConflict
	case apperror.CodeUnauthorized:
		status = http.StatusUnauthorized
	case apperror.CodeForbidden:
		status = http.StatusForbidden
	case apperror.CodeResourceLimit:
		status = http.StatusTooManyRequests
	case apperror.CodeUnavailable:
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, Response{
		Code:      1,
		ErrorCode: string(code),
		Message:   apperror.PublicMessage(err),
	})
}
