package models

import (
    "time"
    "github.com/google/uuid"
)

type ShareRequest struct {
	UserID      string `json:"user_id" validate:"required,uuid"`
	AccessLevel string `json:"access_level" validate:"required,oneof=read write"`
}

type ShareResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AssetInfo represents asset information for manager views
type AssetInfo struct {
	Type        string    `json:"type"` // "folder" or "note"
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	OwnerID     uuid.UUID `json:"owner_id"`
	OwnerName   string    `json:"owner_name"`
	AccessLevel string    `json:"access_level,omitempty"` // only for shared assets
	CreatedAt   time.Time `json:"created_at"`
}