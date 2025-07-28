package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	TeamID    uuid.UUID `json:"team_id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TeamName  string    `json:"team_name" gorm:"not null"`
	CreatedBy uuid.UUID `json:"created_by" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Managers []User `json:"managers" gorm:"many2many:team_managers;joinForeignKey:team_id;joinReferences:manager_id"`
	Members  []User `json:"members" gorm:"many2many:team_members;joinForeignKey:team_id;joinReferences:member_id"`
}

func (Team) TableName() string {
	return "teams"
}

type TeamManager struct {
	TeamID    uuid.UUID `json:"team_id" gorm:"primaryKey"`
	ManagerID uuid.UUID `json:"manager_id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}

func (TeamManager) TableName() string {
	return "team_managers"
}

type TeamMember struct {
	TeamID    uuid.UUID `json:"team_id" gorm:"primaryKey"`
	MemberID  uuid.UUID `json:"member_id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}

func (TeamMember) TableName() string {
	return "team_members"
}
