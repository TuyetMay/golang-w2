package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"asset-management-api/internal/events/types"
	"asset-management-api/pkg/cache"
	"github.com/google/uuid"
)

// CacheEventHandler handles cache invalidation based on Kafka events
type CacheEventHandler struct {
	cacheService cache.CacheService
}

// NewCacheEventHandler creates a new cache event handler
func NewCacheEventHandler(cacheService cache.CacheService) *CacheEventHandler {
	return &CacheEventHandler{
		cacheService: cacheService,
	}
}

// HandleTeamEvent processes team-related events for cache invalidation/updates
func (h *CacheEventHandler) HandleTeamEvent(ctx context.Context, eventData []byte) error {
	// Parse the base event to get event type
	var baseEvent struct {
		EventType string `json:"eventType"`
	}
	
	if err := json.Unmarshal(eventData, &baseEvent); err != nil {
		return fmt.Errorf("failed to parse base team event: %w", err)
	}
	
	switch baseEvent.EventType {
	case types.TeamCreated:
		return h.handleTeamCreated(ctx, eventData)
	case types.MemberAdded:
		return h.handleMemberAdded(ctx, eventData)
	case types.MemberRemoved:
		return h.handleMemberRemoved(ctx, eventData)
	case types.ManagerAdded:
		return h.handleManagerAdded(ctx, eventData)
	case types.ManagerRemoved:
		return h.handleManagerRemoved(ctx, eventData)
	default:
		log.Printf("Unknown team event type: %s", baseEvent.EventType)
		return nil
	}
}

func (h *CacheEventHandler) handleTeamCreated(ctx context.Context, eventData []byte) error {
	var event types.TeamCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse team created event: %w", err)
	}
	
	// Cache initial team members (managers + members)
	allMembers := make([]uuid.UUID, 0, len(event.Managers)+len(event.Members))
	allMembers = append(allMembers, event.Managers...)
	allMembers = append(allMembers, event.Members...)
	
	if err := h.cacheService.CacheTeamMembers(ctx, event.TeamID, allMembers); err != nil {
		log.Printf("Failed to cache team members for team %s: %v", event.TeamID, err)
		return err
	}
	
	log.Printf("Cached team members for new team %s (%d members)", event.TeamID, len(allMembers))
	return nil
}

func (h *CacheEventHandler) handleMemberAdded(ctx context.Context, eventData []byte) error {
	var event types.MemberChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse member added event: %w", err)
	}
	
	// Add member to cache
	if err := h.cacheService.AddTeamMember(ctx, event.TeamID, event.TargetUserID); err != nil {
		log.Printf("Failed to add team member to cache for team %s: %v", event.TeamID, err)
		// Invalidate cache as fallback
		return h.cacheService.InvalidateTeamMembers(ctx, event.TeamID)
	}
	
	log.Printf("Added member %s to team %s cache", event.TargetUserID, event.TeamID)
	return nil
}

func (h *CacheEventHandler) handleMemberRemoved(ctx context.Context, eventData []byte) error {
	var event types.MemberChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse member removed event: %w", err)
	}
	
	// Remove member from cache
	if err := h.cacheService.RemoveTeamMember(ctx, event.TeamID, event.TargetUserID); err != nil {
		log.Printf("Failed to remove team member from cache for team %s: %v", event.TeamID, err)
		// Invalidate cache as fallback
		return h.cacheService.InvalidateTeamMembers(ctx, event.TeamID)
	}
	
	log.Printf("Removed member %s from team %s cache", event.TargetUserID, event.TeamID)
	return nil
}

func (h *CacheEventHandler) handleManagerAdded(ctx context.Context, eventData []byte) error {
	var event types.ManagerChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse manager added event: %w", err)
	}
	
	// Managers are also considered team members for caching purposes
	if err := h.cacheService.AddTeamMember(ctx, event.TeamID, event.TargetUserID); err != nil {
		log.Printf("Failed to add team manager to cache for team %s: %v", event.TeamID, err)
		return h.cacheService.InvalidateTeamMembers(ctx, event.TeamID)
	}
	
	log.Printf("Added manager %s to team %s cache", event.TargetUserID, event.TeamID)
	return nil
}

func (h *CacheEventHandler) handleManagerRemoved(ctx context.Context, eventData []byte) error {
	var event types.ManagerChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse manager removed event: %w", err)
	}
	
	// Remove manager from team members cache
	if err := h.cacheService.RemoveTeamMember(ctx, event.TeamID, event.TargetUserID); err != nil {
		log.Printf("Failed to remove team manager from cache for team %s: %v", event.TeamID, err)
		return h.cacheService.InvalidateTeamMembers(ctx, event.TeamID)
	}
	
	log.Printf("Removed manager %s from team %s cache", event.TargetUserID, event.TeamID)
	return nil
}

// HandleAssetEvent processes asset-related events for cache invalidation/updates
func (h *CacheEventHandler) HandleAssetEvent(ctx context.Context, eventData []byte) error {
	// Parse the base event to get event type
	var baseEvent struct {
		EventType string `json:"eventType"`
		AssetType string `json:"assetType"`
	}
	
	if err := json.Unmarshal(eventData, &baseEvent); err != nil {
		return fmt.Errorf("failed to parse base asset event: %w", err)
	}
	
	switch baseEvent.EventType {
	case types.FolderCreated, types.NoteCreated:
		return h.handleAssetCreated(ctx, eventData, baseEvent.AssetType)
	case types.FolderUpdated, types.NoteUpdated:
		return h.handleAssetUpdated(ctx, eventData, baseEvent.AssetType)
	case types.FolderDeleted, types.NoteDeleted:
		return h.handleAssetDeleted(ctx, eventData, baseEvent.AssetType)
	case types.FolderShared, types.NoteShared:
		return h.handleAssetShared(ctx, eventData, baseEvent.AssetType)
	case types.FolderUnshared, types.NoteUnshared:
		return h.handleAssetUnshared(ctx, eventData, baseEvent.AssetType)
	default:
		log.Printf("Unknown asset event type: %s", baseEvent.EventType)
		return nil
	}
}

func (h *CacheEventHandler) handleAssetCreated(ctx context.Context, eventData []byte, assetType string) error {
	var event types.AssetCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse asset created event: %w", err)
	}
	
	// No need to cache on creation, cache will be populated on first read
	log.Printf("Asset %s (%s) created: %s", assetType, event.AssetID, event.Name)
	return nil
}

func (h *CacheEventHandler) handleAssetUpdated(ctx context.Context, eventData []byte, assetType string) error {
	var event types.AssetUpdatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse asset updated event: %w", err)
	}
	
	// Invalidate metadata cache since asset was updated
	if assetType == types.AssetTypeFolder {
		if err := h.cacheService.InvalidateFolderMetadata(ctx, event.AssetID); err != nil {
			log.Printf("Failed to invalidate folder metadata cache for %s: %v", event.AssetID, err)
		}
	} else if assetType == types.AssetTypeNote {
		if err := h.cacheService.InvalidateNoteMetadata(ctx, event.AssetID); err != nil {
			log.Printf("Failed to invalidate note metadata cache for %s: %v", event.AssetID, err)
		}
	}
	
	log.Printf("Invalidated metadata cache for %s %s due to update", assetType, event.AssetID)
	return nil
}

func (h *CacheEventHandler) handleAssetDeleted(ctx context.Context, eventData []byte, assetType string) error {
	var event types.AssetDeletedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse asset deleted event: %w", err)
	}
	
	// Invalidate all caches related to this asset
	if assetType == types.AssetTypeFolder {
		if err := h.cacheService.InvalidateFolderMetadata(ctx, event.AssetID); err != nil {
			log.Printf("Failed to invalidate folder metadata cache for %s: %v", event.AssetID, err)
		}
	} else if assetType == types.AssetTypeNote {
		if err := h.cacheService.InvalidateNoteMetadata(ctx, event.AssetID); err != nil {
			log.Printf("Failed to invalidate note metadata cache for %s: %v", event.AssetID, err)
		}
	}
	
	// Invalidate ACL cache
	if err := h.cacheService.InvalidateAssetACL(ctx, event.AssetID); err != nil {
		log.Printf("Failed to invalidate asset ACL cache for %s: %v", event.AssetID, err)
	}
	
	log.Printf("Invalidated all caches for deleted %s %s", assetType, event.AssetID)
	return nil
}

func (h *CacheEventHandler) handleAssetShared(ctx context.Context, eventData []byte, assetType string) error {
	var event types.AssetSharedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse asset shared event: %w", err)
	}
	
	// Update ACL cache
	if err := h.cacheService.UpdateAssetACL(ctx, event.AssetID, event.SharedWithUserID, event.AccessLevel); err != nil {
		log.Printf("Failed to update asset ACL cache for %s: %v", event.AssetID, err)
		// Invalidate ACL cache as fallback
		if err := h.cacheService.InvalidateAssetACL(ctx, event.AssetID); err != nil {
			log.Printf("Failed to invalidate asset ACL cache for %s: %v", event.AssetID, err)
		}
	}
	
	log.Printf("Updated ACL cache for %s %s: user %s granted %s access", 
		assetType, event.AssetID, event.SharedWithUserID, event.AccessLevel)
	return nil
}

func (h *CacheEventHandler) handleAssetUnshared(ctx context.Context, eventData []byte, assetType string) error {
	var event types.AssetUnsharedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return fmt.Errorf("failed to parse asset unshared event: %w", err)
	}
	
	// Remove user from ACL cache
	if err := h.cacheService.RemoveAssetACL(ctx, event.AssetID, event.UnsharedFromUserID); err != nil {
		log.Printf("Failed to remove user from asset ACL cache for %s: %v", event.AssetID, err)
		// Invalidate ACL cache as fallback
		if err := h.cacheService.InvalidateAssetACL(ctx, event.AssetID); err != nil {
			log.Printf("Failed to invalidate asset ACL cache for %s: %v", event.AssetID, err)
		}
	}
	
	log.Printf("Removed user %s from ACL cache for %s %s", 
		event.UnsharedFromUserID, assetType, event.AssetID)
	return nil
}