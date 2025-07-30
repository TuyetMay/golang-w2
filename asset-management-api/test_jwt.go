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
	// Dùng same secret như trong .env
	secretKey := "your-super-secret-jwt-key-change-in-production-make-it-long-and-random"
	
	// Create a manager user
	managerID := uuid.New()
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
	
	// Create a member user
	memberID := uuid.New()
	memberClaims := &Claims{
		UserID:   memberID,
		Email:    "member@test.com",
		Role:     "member",
		Username: "test_member",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   memberID.String(),
		},
	}

	memberToken := jwt.NewWithClaims(jwt.SigningMethodHS256, memberClaims)
	memberTokenString, err := memberToken.SignedString([]byte(secretKey))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== MEMBER TOKEN ===")
	fmt.Println(memberTokenString)
	fmt.Println("\nUserID:", memberID.String())
}