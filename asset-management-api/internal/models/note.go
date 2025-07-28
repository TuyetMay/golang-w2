package models

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	NoteID    uuid.UUID `json:"note_id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title     string    `json:"title" gorm:"not null"`
	Body      string    `json:"body"`
	FolderID  uuid.UUID `json:"folder_id" gorm:"not null"`
	OwnerID   uuid.UUID `json:"owner_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Folder Folder `json:"folder" gorm:"foreignKey:FolderID"`
	Owner  User   `json:"owner" gorm:"foreignKey:OwnerID"`
}

func (Note) TableName() string {
	return "notes"
}

type NoteShare struct {
	NoteID           uuid.UUID `json:"note_id" gorm:"primaryKey"`
	SharedWithUserID uuid.UUID `json:"shared_with_user_id" gorm:"primaryKey"`
	AccessLevel      string    `json:"access_level" gorm:"not null;check:access_level IN ('read','write')"`
	SharedBy         uuid.UUID `json:"shared_by" gorm:"not null"`
	CreatedAt        time.Time `json:"created_at"`

	// Relationships
	Note           Note `json:"note" gorm:"foreignKey:NoteID"`
	SharedWithUser User `json:"shared_with_user" gorm:"foreignKey:SharedWithUserID"`
	SharedByUser   User `json:"shared_by_user" gorm:"foreignKey:SharedBy"`
}

func (NoteShare) TableName() string {
	return "note_shares"
}
