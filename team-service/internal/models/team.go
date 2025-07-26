package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Team struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"not null"`
	CreatedBy uuid.UUID      `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	Creator  User           `json:"creator" gorm:"foreignKey:CreatedBy"`
	Members  []TeamMember   `json:"members,omitempty" gorm:"foreignKey:TeamID"`
	Managers []TeamManager  `json:"managers,omitempty" gorm:"foreignKey:TeamID"`
}

// Request/Response DTOs
type CreateTeamRequest struct {
	TeamName string            `json:"teamName" binding:"required"`
	Managers []TeamMemberInfo  `json:"managers,omitempty"`
	Members  []TeamMemberInfo  `json:"members,omitempty"`
}

type TeamMemberInfo struct {
	ID   string `json:"managerId,omitempty"`
	Name string `json:"managerName,omitempty"`
}

type AddMemberRequest struct {
	UserID string `json:"userId" binding:"required"`
}

type TeamResponse struct {
	ID        uuid.UUID        `json:"id"`
	Name      string           `json:"name"`
	CreatedBy uuid.UUID        `json:"createdBy"`
	CreatedAt time.Time        `json:"createdAt"`
	Members   []UserInfo       `json:"members"`
	Managers  []UserInfo       `json:"managers"`
}

type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
}