package models

import (
	"time"

	"github.com/google/uuid"
)

type Folder struct {
	FolderID    uuid.UUID `json:"folder_id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Owner User   `json:"owner" gorm:"foreignKey:OwnerID"`
	Notes []Note `json:"notes,omitempty" gorm:"foreignKey:FolderID"`
}

func (Folder) TableName() string {
	return "folders"
}

type FolderShare struct {
	FolderID         uuid.UUID `json:"folder_id" gorm:"primaryKey"`
	SharedWithUserID uuid.UUID `json:"shared_with_user_id" gorm:"primaryKey"`
	AccessLevel      string    `json:"access_level" gorm:"not null;check:access_level IN ('read','write')"`
	SharedBy         uuid.UUID `json:"shared_by" gorm:"not null"`
	CreatedAt        time.Time `json:"created_at"`

	// Relationships
	Folder         Folder `json:"folder" gorm:"foreignKey:FolderID"`
	SharedWithUser User   `json:"shared_with_user" gorm:"foreignKey:SharedWithUserID"`
	SharedByUser   User   `json:"shared_by_user" gorm:"foreignKey:SharedBy"`
}

func (FolderShare) TableName() string {
	return "folder_shares"
}
