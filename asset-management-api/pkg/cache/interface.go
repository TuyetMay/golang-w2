package cache

import (
	"context"
	"time"

	"github.com/google/uuid"
	"asset-management-api/internal/models"
)

// CacheService defines the interface for caching operations
type CacheService interface {
	// Team member caching
	CacheTeamMembers(ctx context.Context, teamID uuid.UUID, members []uuid.UUID) error
	GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]uuid.UUID, error)
	AddTeamMember(ctx context.Context, teamID, memberID uuid.UUID) error
	RemoveTeamMember(ctx context.Context, teamID, memberID uuid.UUID) error
	InvalidateTeamMembers(ctx context.Context, teamID uuid.UUID) error

	// Asset metadata caching
	CacheFolderMetadata(ctx context.Context, folder *models.Folder) error
	GetFolderMetadata(ctx context.Context, folderID uuid.UUID) (*models.Folder, error)
	CacheNoteMetadata(ctx context.Context, note *models.Note) error
	GetNoteMetadata(ctx context.Context, noteID uuid.UUID) (*models.Note, error)
	InvalidateFolderMetadata(ctx context.Context, folderID uuid.UUID) error
	InvalidateNoteMetadata(ctx context.Context, noteID uuid.UUID) error

	// Access control caching
	CacheAssetACL(ctx context.Context, assetID uuid.UUID, acl map[string]string) error
	GetAssetACL(ctx context.Context, assetID uuid.UUID) (map[string]string, error)
	UpdateAssetACL(ctx context.Context, assetID, userID uuid.UUID, accessLevel string) error
	RemoveAssetACL(ctx context.Context, assetID, userID uuid.UUID) error
	InvalidateAssetACL(ctx context.Context, assetID uuid.UUID) error

	// Generic cache operations
	HealthCheck() map[string]interface{}
	Close() error
}

// EventHandler defines the interface for handling cache invalidation events
type EventHandler interface {
	HandleTeamEvent(ctx context.Context, eventData []byte) error
	HandleAssetEvent(ctx context.Context, eventData []byte) error
}

// CacheKeys defines standard cache key formats
type CacheKeys struct{}

func (CacheKeys) TeamMembers(teamID uuid.UUID) string {
	return "team:" + teamID.String() + ":members"
}

func (CacheKeys) FolderMetadata(folderID uuid.UUID) string {
	return "folder:" + folderID.String()
}

func (CacheKeys) NoteMetadata(noteID uuid.UUID) string {
	return "note:" + noteID.String()
}

func (CacheKeys) AssetACL(assetID uuid.UUID) string {
	return "asset:" + assetID.String() + ":acl"
}

// Default cache TTL values
const (
	DefaultTeamMembersTTL = 1 * time.Hour
	DefaultAssetTTL       = 30 * time.Minute
	DefaultACLTTL         = 15 * time.Minute
)