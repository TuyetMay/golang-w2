package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string, err string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error:   err,
	})
}

func ValidationErrorResponse(c *gin.Context, errors []string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": "Validation failed",
		"errors":  errors,
	})
}

func PaginatedSuccessResponse(c *gin.Context, statusCode int, message string, data interface{}, pagination *Pagination) {
	c.JSON(statusCode, PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}

func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message, "Authentication required")
}

func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message, "Access denied")
}

func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message, "Resource not found")
}

func InternalServerErrorResponse(c *gin.Context, message string, err error) {
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}
	ErrorResponse(c, http.StatusInternalServerError, message, errorMsg)
}

func BadRequestResponse(c *gin.Context, message string, err error) {
	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	}
	ErrorResponse(c, http.StatusBadRequest, message, errorMsg)
}