package types

import (
	"time"
	"github.com/google/uuid"
)

// Asset event types
const (
	// Folder events
	FolderCreated = "FOLDER_CREATED"
	FolderUpdated = "FOLDER_UPDATED"
	FolderDeleted = "FOLDER_DELETED"
	FolderShared  = "FOLDER_SHARED"
	FolderUnshared = "FOLDER_UNSHARED"
	
	// Note events
	NoteCreated   = "NOTE_CREATED"
	NoteUpdated   = "NOTE_UPDATED"
	NoteDeleted   = "NOTE_DELETED"
	NoteShared    = "NOTE_SHARED"
	NoteUnshared  = "NOTE_UNSHARED"
)

// Asset types
const (
	AssetTypeFolder = "folder"
	AssetTypeNote   = "note"
)

// Topics
const (
	AssetChangesTopic = "asset.changes"
)

// BaseAssetEvent represents the common fields for all asset events
type BaseAssetEvent struct {
	EventType string    `json:"eventType"`
	AssetType string    `json:"assetType"`
	AssetID   uuid.UUID `json:"assetId"`
	OwnerID   uuid.UUID `json:"ownerId"`
	ActionBy  uuid.UUID `json:"actionBy"`
	Timestamp time.Time `json:"timestamp"`
}

// AssetCreatedEvent represents asset creation events
type AssetCreatedEvent struct {
	BaseAssetEvent
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	FolderID    uuid.UUID `json:"folderId,omitempty"` // Only for notes
}

// AssetUpdatedEvent represents asset update events
type AssetUpdatedEvent struct {
	BaseAssetEvent
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Changes     []string  `json:"changes"` // List of changed fields
}

// AssetDeletedEvent represents asset deletion events
type AssetDeletedEvent struct {
	BaseAssetEvent
	Name string `json:"name"`
}

// AssetSharedEvent represents asset sharing events
type AssetSharedEvent struct {
	BaseAssetEvent
	SharedWithUserID uuid.UUID `json:"sharedWithUserId"`
	AccessLevel      string    `json:"accessLevel"`
	SharedByUserName string    `json:"sharedByUserName"`
}

// AssetUnsharedEvent represents asset unsharing events
type AssetUnsharedEvent struct {
	BaseAssetEvent
	UnsharedFromUserID uuid.UUID `json:"unsharedFromUserId"`
	UnsharedByUserName string    `json:"unsharedByUserName"`
}

// Constructor functions for folder events
func NewFolderCreatedEvent(folderID, ownerID, actionBy uuid.UUID, name, description string) *AssetCreatedEvent {
	return &AssetCreatedEvent{
		BaseAssetEvent: BaseAssetEvent{
			EventType: FolderCreated,
			AssetType: AssetTypeFolder,
			AssetID:   folderID,
			OwnerID:   ownerID,
			ActionBy:  actionBy,
			Timestamp: time.Now().UTC(),
		},
		Name:        name,
		Description: description,
	}
}

func NewFolderUpdatedEvent(folderID, ownerID, actionBy uuid.UUID, name, description string, changes []string) *AssetUpdatedEvent {
	return &AssetUpdatedEvent{
		BaseAssetEvent: BaseAssetEvent{
			EventType: FolderUpdated,
			AssetType: AssetTypeFolder,
			AssetID:   folderID,
			OwnerID:   ownerID,
			ActionBy:  actionBy,
			Timestamp: time.Now().UTC(),
		},
		Name:        name,
		Description: description,
		Changes:     changes,
	}
}

func NewFolderDeletedEvent(folderID, ownerID, actionBy uuid.UUID, name string) *AssetDeletedEvent {
	return &AssetDeletedEvent{
		BaseAssetEvent: BaseAssetEvent{
			EventType: FolderDeleted,
			AssetType: AssetTypeFolder,
			AssetID:   folderID,
			OwnerID:   ownerID,
			ActionBy:  actionBy,
			Timestamp: time.Now().UTC(),
		},
		Name: name,
	}
}

// Constructor functions for note events
func NewNoteCreatedEvent(noteID, ownerID, actionBy, folderID uuid.UUID, title, body string) *AssetCreatedEvent {
	return &AssetCreatedEvent{
		BaseAssetEvent: BaseAssetEvent{
			EventType: NoteCreated,
			AssetType: AssetTypeNote,
			AssetID:   noteID,
			OwnerID:   ownerID,
			ActionBy:  actionBy,
			Timestamp: time.Now().UTC(),
		},
		Name:        title,
		Description: body,
		FolderID:    folderID,
	}
}

func NewNoteUpdatedEvent(noteID, ownerID, actionBy uuid.UUID, title, body string, changes []string) *AssetUpdatedEvent {
	return &AssetUpdatedEvent{
		BaseAssetEvent: BaseAssetEvent{
			EventType: NoteUpdated,
			AssetType: AssetTypeNote,
			AssetID:   noteID,
			OwnerID:   ownerID,
			ActionBy:  actionBy,
			Timestamp: time.Now().UTC(),
		},
		Name:        title,
		Description: body,
		Changes:     changes,
	}
}

func NewNoteDeletedEvent(noteID, ownerID, actionBy uuid.UUID, title string) *AssetDeletedEvent {
	return &AssetDeletedEvent{
		BaseAssetEvent: BaseAssetEvent{
			EventType: NoteDeleted,
			AssetType: AssetTypeNote,
			AssetID:   noteID,
			OwnerID:   ownerID,
			ActionBy:  actionBy,
			Timestamp: time.Now().UTC(),
		},
		Name: title,
	}
}

// Constructor functions for sharing events
func NewAssetSharedEvent(eventType, assetType string, assetID, ownerID, actionBy, sharedWithUserID uuid.UUID, accessLevel, sharedByUserName string) *AssetSharedEvent {
	return &AssetSharedEvent{
		BaseAssetEvent: BaseAssetEvent{
			EventType: eventType,
			AssetType: assetType,
			AssetID:   assetID,
			OwnerID:   ownerID,
			ActionBy:  actionBy,
			Timestamp: time.Now().UTC(),
		},
		SharedWithUserID: sharedWithUserID,
		AccessLevel:      accessLevel,
		SharedByUserName: sharedByUserName,
	}
}

func NewAssetUnsharedEvent(eventType, assetType string, assetID, ownerID, actionBy, unsharedFromUserID uuid.UUID, unsharedByUserName string) *AssetUnsharedEvent {
	return &AssetUnsharedEvent{
		BaseAssetEvent: BaseAssetEvent{
			EventType: eventType,
			AssetType: assetType,
			AssetID:   assetID,
			OwnerID:   ownerID,
			ActionBy:  actionBy,
			Timestamp: time.Now().UTC(),
		},
		UnsharedFromUserID: unsharedFromUserID,
		UnsharedByUserName: unsharedByUserName,
	}
}