package middleware

import (
	"asset-management-api/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	jwtUtil *utils.JWTUtil
}

func NewAuthMiddleware(jwtUtil *utils.JWTUtil) *AuthMiddleware {
	return &AuthMiddleware{jwtUtil: jwtUtil}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.UnauthorizedResponse(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtUtil.ValidateToken(tokenString)
		if err != nil {
			utils.UnauthorizedResponse(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("username", claims.Username)
		c.Set("claims", claims)

		c.Next()
	}
}

func (m *AuthMiddleware) RequireManagerRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware should be used after RequireAuth
		userRole, exists := c.Get("user_role")
		if !exists {
			utils.UnauthorizedResponse(c, "Authentication required")
			c.Abort()
			return
		}

		if userRole != "manager" {
			utils.ForbiddenResponse(c, "Manager role required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helper function to get user ID from context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	
	if id, ok := userID.(uuid.UUID); ok {
		return id, true
	}
	
	return uuid.Nil, false
}

// Helper function to get user role from context
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	
	if role, ok := userRole.(string); ok {
		return role, true
	}
	
	return "", false
}

// Helper function to get username from context
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	
	if name, ok := username.(string); ok {
		return name, true
	}
	
	return "", false
}
