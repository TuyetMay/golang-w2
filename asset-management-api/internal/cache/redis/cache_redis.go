package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"asset-management-api/internal/models"
	"asset-management-api/pkg/cache"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisCacheService implements the CacheService interface using Redis
type RedisCacheService struct {
	client *RedisClient
	keys   cache.CacheKeys
}

// NewRedisCacheService creates a new Redis cache service
func NewRedisCacheService(client *RedisClient) *RedisCacheService {
	return &RedisCacheService{
		client: client,
		keys:   cache.CacheKeys{},
	}
}

// Team member caching methods
func (r *RedisCacheService) CacheTeamMembers(ctx context.Context, teamID uuid.UUID, members []uuid.UUID) error {
	key := r.keys.TeamMembers(teamID)
	
	// Clear existing members
	if err := r.client.Del(ctx, key); err != nil {
		log.Printf("Warning: failed to clear existing team members cache: %v", err)
	}
	
	// Add all members to list
	if len(members) > 0 {
		memberStrs := make([]interface{}, len(members))
		for i, member := range members {
			memberStrs[i] = member.String()
		}
		
		if err := r.client.LPush(ctx, key, memberStrs...); err != nil {
			return fmt.Errorf("failed to cache team members: %w", err)
		}
		
		// Set expiration
		if err := r.client.Expire(ctx, key, cache.DefaultTeamMembersTTL); err != nil {
			log.Printf("Warning: failed to set expiration for team members cache: %v", err)
		}
	}
	
	return nil
}

func (r *RedisCacheService) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]uuid.UUID, error) {
	key := r.keys.TeamMembers(teamID)
	
	memberStrs, err := r.client.LRange(ctx, key, 0, -1)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get team members from cache: %w", err)
	}
	
	members := make([]uuid.UUID, len(memberStrs))
	for i, memberStr := range memberStrs {
		memberID, err := uuid.Parse(memberStr)
		if err != nil {
			log.Printf("Warning: invalid UUID in team members cache: %s", memberStr)
			continue
		}
		members[i] = memberID
	}
	
	return members, nil
}

func (r *RedisCacheService) AddTeamMember(ctx context.Context, teamID, memberID uuid.UUID) error {
	key := r.keys.TeamMembers(teamID)
	
	// Check if key exists, if not, skip (cache will be populated on next read)
	exists, err := r.client.Exists(ctx, key)
	if err != nil || !exists {
		return nil
	}
	
	// Add member to list
	if err := r.client.LPush(ctx, key, memberID.String()); err != nil {
		return fmt.Errorf("failed to add team member to cache: %w", err)
	}
	
	return nil
}

func (r *RedisCacheService) RemoveTeamMember(ctx context.Context, teamID, memberID uuid.UUID) error {
	key := r.keys.TeamMembers(teamID)
	
	// Remove all occurrences of the member
	if err := r.client.LRem(ctx, key, 0, memberID.String()); err != nil {
		if !errors.Is(err, redis.Nil) {
			return fmt.Errorf("failed to remove team member from cache: %w", err)
		}
	}
	
	return nil
}

func (r *RedisCacheService) InvalidateTeamMembers(ctx context.Context, teamID uuid.UUID) error {
	key := r.keys.TeamMembers(teamID)
	return r.client.Del(ctx, key)
}

// Asset metadata caching methods
func (r *RedisCacheService) CacheFolderMetadata(ctx context.Context, folder *models.Folder) error {
	key := r.keys.FolderMetadata(folder.FolderID)
	
	if err := r.client.SetJSON(ctx, key, folder, cache.DefaultAssetTTL); err != nil {
		return fmt.Errorf("failed to cache folder metadata: %w", err)
	}
	
	return nil
}

func (r *RedisCacheService) GetFolderMetadata(ctx context.Context, folderID uuid.UUID) (*models.Folder, error) {
	key := r.keys.FolderMetadata(folderID)
	
	var folder models.Folder
	err := r.client.GetJSON(ctx, key, &folder)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get folder metadata from cache: %w", err)
	}
	
	return &folder, nil
}

func (r *RedisCacheService) CacheNoteMetadata(ctx context.Context, note *models.Note) error {
	key := r.keys.NoteMetadata(note.NoteID)
	
	if err := r.client.SetJSON(ctx, key, note, cache.DefaultAssetTTL); err != nil {
		return fmt.Errorf("failed to cache note metadata: %w", err)
	}
	
	return nil
}

func (r *RedisCacheService) GetNoteMetadata(ctx context.Context, noteID uuid.UUID) (*models.Note, error) {
	key := r.keys.NoteMetadata(noteID)
	
	var note models.Note
	err := r.client.GetJSON(ctx, key, &note)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get note metadata from cache: %w", err)
	}
	
	return &note, nil
}

func (r *RedisCacheService) InvalidateFolderMetadata(ctx context.Context, folderID uuid.UUID) error {
	key := r.keys.FolderMetadata(folderID)
	return r.client.Del(ctx, key)
}

func (r *RedisCacheService) InvalidateNoteMetadata(ctx context.Context, noteID uuid.UUID) error {
	key := r.keys.NoteMetadata(noteID)
	return r.client.Del(ctx, key)
}

// Access control caching methods
func (r *RedisCacheService) CacheAssetACL(ctx context.Context, assetID uuid.UUID, acl map[string]string) error {
	key := r.keys.AssetACL(assetID)
	
	// Convert map to interface{} slice for HSet
	fields := make([]interface{}, 0, len(acl)*2)
	for userID, accessLevel := range acl {
		fields = append(fields, userID, accessLevel)
	}
	
	if len(fields) > 0 {
		if err := r.client.HSet(ctx, key, fields...); err != nil {
			return fmt.Errorf("failed to cache asset ACL: %w", err)
		}
		
		// Set expiration
		if err := r.client.Expire(ctx, key, cache.DefaultACLTTL); err != nil {
			log.Printf("Warning: failed to set expiration for asset ACL cache: %v", err)
		}
	}
	
	return nil
}

func (r *RedisCacheService) GetAssetACL(ctx context.Context, assetID uuid.UUID) (map[string]string, error) {
	key := r.keys.AssetACL(assetID)
	
	acl, err := r.client.HGetAll(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get asset ACL from cache: %w", err)
	}
	
	if len(acl) == 0 {
		return nil, nil // Cache miss or empty ACL
	}
	
	return acl, nil
}

func (r *RedisCacheService) UpdateAssetACL(ctx context.Context, assetID, userID uuid.UUID, accessLevel string) error {
	key := r.keys.AssetACL(assetID)
	
	// Check if key exists
	exists, err := r.client.Exists(ctx, key)
	if err != nil || !exists {
		return nil // Cache doesn't exist, skip update
	}
	
	// Update the specific user's access level
	if err := r.client.HSet(ctx, key, userID.String(), accessLevel); err != nil {
		return fmt.Errorf("failed to update asset ACL in cache: %w", err)
	}
	
	return nil
}

func (r *RedisCacheService) RemoveAssetACL(ctx context.Context, assetID, userID uuid.UUID) error {
	key := r.keys.AssetACL(assetID)
	
	if err := r.client.HDel(ctx, key, userID.String()); err != nil {
		if !errors.Is(err, redis.Nil) {
			return fmt.Errorf("failed to remove asset ACL from cache: %w", err)
		}
	}
	
	return nil
}

func (r *RedisCacheService) InvalidateAssetACL(ctx context.Context, assetID uuid.UUID) error {
	key := r.keys.AssetACL(assetID)
	return r.client.Del(ctx, key)
}

// Health check and cleanup
func (r *RedisCacheService) HealthCheck() map[string]interface{} {
	return r.client.Health()
}

func (r *RedisCacheService) Close() error {
	return r.client.Close()
}