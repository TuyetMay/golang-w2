// File: test_jwt.go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	secretKey := "your-super-secret-jwt-key-change-in-production-make-it-long-and-random"
	
	// Use existing user ID from database
	managerID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	
	claims := &Claims{
		UserID:   managerID,
		Email:    "manager@test.com",
		Role:     "manager", 
		Username: "test_manager",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   managerID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== MANAGER TOKEN ===")
	fmt.Println(tokenString)
	fmt.Println("\nUserID:", managerID.String())
}