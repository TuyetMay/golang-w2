package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamMember struct {
	ID       uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID   uuid.UUID      `json:"team_id" gorm:"type:uuid;not null"`
	UserID   uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	AddedBy  uuid.UUID      `json:"added_by" gorm:"type:uuid;not null"`
	AddedAt  time.Time      `json:"added_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type TeamManager struct {
	ID       uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TeamID   uuid.UUID      `json:"team_id" gorm:"type:uuid;not null"`
	UserID   uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	AddedBy  uuid.UUID      `json:"added_by" gorm:"type:uuid;not null"`
	AddedAt  time.Time      `json:"added_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relationships
	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}