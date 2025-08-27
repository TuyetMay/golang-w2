package service

import (
	"context"
	"log"

	"asset-management-api/internal/models"
	"asset-management-api/pkg/cache"
	"github.com/google/uuid"
)

// CacheIntegratedFolderService wraps the folder service with caching capabilities
type CacheIntegratedFolderService struct {
	folderService FolderService
	cacheService  cache.CacheService
}

// NewCacheIntegratedFolderService creates a new cache-integrated folder service
func NewCacheIntegratedFolderService(folderService FolderService, cacheService cache.CacheService) *CacheIntegratedFolderService {
	return &CacheIntegratedFolderService{
		folderService: folderService,
		cacheService:  cacheService,
	}
}

// GetFolder attempts to get folder from cache first, then falls back to database
func (s *CacheIntegratedFolderService) GetFolder(folderID, userID uuid.UUID) (*models.Folder, error) {
	ctx := context.Background()
	
	// Try to get from cache first
	if cachedFolder, err := s.cacheService.GetFolderMetadata(ctx, folderID); err == nil && cachedFolder != nil {
		log.Printf("Cache HIT for folder %s", folderID)
		return cachedFolder, nil
	}
	
	log.Printf("Cache MISS for folder %s, fetching from database", folderID)
	
	// Cache miss, get from database
	folder, err := s.folderService.GetFolder(folderID, userID)
	if err != nil {
		return nil, err
	}
	
	// Cache the result for future requests
	if err := s.cacheService.CacheFolderMetadata(ctx, folder); err != nil {
		log.Printf("Failed to cache folder metadata for %s: %v", folderID, err)
	}
	
	return folder, nil
}

// CreateFolder creates folder and caches it
func (s *CacheIntegratedFolderService) CreateFolder(userID uuid.UUID, name, description string) (*models.Folder, error) {
	folder, err := s.folderService.CreateFolder(userID, name, description)
	if err != nil {
		return nil, err
	}
	
	// Cache the newly created folder
	ctx := context.Background()
	if err := s.cacheService.CacheFolderMetadata(ctx, folder); err != nil {
		log.Printf("Failed to cache newly created folder %s: %v", folder.FolderID, err)
	}
	
	return folder, nil
}

// UpdateFolder updates folder and invalidates cache
func (s *CacheIntegratedFolderService) UpdateFolder(folderID, userID uuid.UUID, name, description string) (*models.Folder, error) {
	folder, err := s.folderService.UpdateFolder(folderID, userID, name, description)
	if err != nil {
		return nil, err
	}
	
	// Cache invalidation is handled by Kafka event handler
	// but we can also update the cache directly for immediate consistency
	ctx := context.Background()
	if err := s.cacheService.CacheFolderMetadata(ctx, folder); err != nil {
		log.Printf("Failed to cache updated folder %s: %v", folder.FolderID, err)
	}
	
	return folder, nil
}

// DeleteFolder deletes folder and invalidates cache
func (s *CacheIntegratedFolderService) DeleteFolder(folderID, userID uuid.UUID) error {
	err := s.folderService.DeleteFolder(folderID, userID)
	if err != nil {
		return err
	}
	
	// Cache invalidation is handled by Kafka event handler
	return nil
}

// GetUserFolders gets user folders with caching support
func (s *CacheIntegratedFolderService) GetUserFolders(userID uuid.UUID) ([]*models.Folder, error) {
	// For list operations, we typically don't cache the entire list
	// Instead, individual folder metadata will be cached when accessed
	return s.folderService.GetUserFolders(userID)
}

// CacheIntegratedNoteService wraps the note service with caching capabilities
type CacheIntegratedNoteService struct {
	noteService  NoteService
	cacheService cache.CacheService
}

// NewCacheIntegratedNoteService creates a new cache-integrated note service
func NewCacheIntegratedNoteService(noteService NoteService, cacheService cache.CacheService) *CacheIntegratedNoteService {
	return &CacheIntegratedNoteService{
		noteService:  noteService,
		cacheService: cacheService,
	}
}

// GetNote attempts to get note from cache first, then falls back to database
func (s *CacheIntegratedNoteService) GetNote(noteID, userID uuid.UUID) (*models.Note, error) {
	ctx := context.Background()
	
	// Try to get from cache first
	if cachedNote, err := s.cacheService.GetNoteMetadata(ctx, noteID); err == nil && cachedNote != nil {
		log.Printf("Cache HIT for note %s", noteID)
		return cachedNote, nil
	}
	
	log.Printf("Cache MISS for note %s, fetching from database", noteID)
	
	// Cache miss, get from database
	note, err := s.noteService.GetNote(noteID, userID)
	if err != nil {
		return nil, err
	}
	
	// Cache the result for future requests
	if err := s.cacheService.CacheNoteMetadata(ctx, note); err != nil {
		log.Printf("Failed to cache note metadata for %s: %v", noteID, err)
	}
	
	return note, nil
}

// CreateNote creates note and caches it
func (s *CacheIntegratedNoteService) CreateNote(userID, folderID uuid.UUID, title, body string) (*models.Note, error) {
	note, err := s.noteService.CreateNote(userID, folderID, title, body)
	if err != nil {
		return nil, err
	}
	
	// Cache the newly created note
	ctx := context.Background()
	if err := s.cacheService.CacheNoteMetadata(ctx, note); err != nil {
		log.Printf("Failed to cache newly created note %s: %v", note.NoteID, err)
	}
	
	return note, nil
}

// UpdateNote updates note and refreshes cache
func (s *CacheIntegratedNoteService) UpdateNote(noteID, userID uuid.UUID, title, body string) (*models.Note, error) {
	note, err := s.noteService.UpdateNote(noteID, userID, title, body)
	if err != nil {
		return nil, err
	}
	
	// Update cache with new data
	ctx := context.Background()
	if err := s.cacheService.CacheNoteMetadata(ctx, note); err != nil {
		log.Printf("Failed to cache updated note %s: %v", note.NoteID, err)
	}
	
	return note, nil
}

// DeleteNote deletes note and invalidates cache
func (s *CacheIntegratedNoteService) DeleteNote(noteID, userID uuid.UUID) error {
	err := s.noteService.DeleteNote(noteID, userID)
	if err != nil {
		return err
	}
	
	// Cache invalidation is handled by Kafka event handler
	return nil
}

// GetNotesByFolder gets notes by folder
func (s *CacheIntegratedNoteService) GetNotesByFolder(folderID, userID uuid.UUID) ([]*models.Note, error) {
	// For list operations, we typically don't cache the entire list
	return s.noteService.GetNotesByFolder(folderID, userID)
}

// GetUserNotes gets user notes
func (s *CacheIntegratedNoteService) GetUserNotes(userID uuid.UUID) ([]*models.Note, error) {
	return s.noteService.GetUserNotes(userID)
}

// CacheIntegratedTeamService wraps the team service with caching capabilities
type CacheIntegratedTeamService struct {
	teamService  TeamService
	cacheService cache.CacheService
}

// NewCacheIntegratedTeamService creates a new cache-integrated team service
func NewCacheIntegratedTeamService(teamService TeamService, cacheService cache.CacheService) *CacheIntegratedTeamService {
	return &CacheIntegratedTeamService{
		teamService:  teamService,
		cacheService: cacheService,
	}
}

// CreateTeam creates team and caches members
func (s *CacheIntegratedTeamService) CreateTeam(creatorID uuid.UUID, teamName string, managers []TeamMemberInfo, members []TeamMemberInfo) (*models.Team, error) {
	team, err := s.teamService.CreateTeam(creatorID, teamName, managers, members)
	if err != nil {
		return nil, err
	}
	
	// Cache team members will be handled by Kafka event handler
	return team, nil
}

// AddMember adds member to team and updates cache
func (s *CacheIntegratedTeamService) AddMember(teamID, requestorID, memberID uuid.UUID) error {
	err := s.teamService.AddMember(teamID, requestorID, memberID)
	if err != nil {
		return err
	}
	
	// Cache update is handled by Kafka event handler
	return nil
}

// RemoveMember removes member from team and updates cache
func (s *CacheIntegratedTeamService) RemoveMember(teamID, requestorID, memberID uuid.UUID) error {
	err := s.teamService.RemoveMember(teamID, requestorID, memberID)
	if err != nil {
		return err
	}
	
	// Cache update is handled by Kafka event handler
	return nil
}

// AddManager adds manager to team and updates cache
func (s *CacheIntegratedTeamService) AddManager(teamID, requestorID, managerID uuid.UUID) error {
	err := s.teamService.AddManager(teamID, requestorID, managerID)
	if err != nil {
		return err
	}
	
	// Cache update is handled by Kafka event handler
	return nil
}

// RemoveManager removes manager from team and updates cache
func (s *CacheIntegratedTeamService) RemoveManager(teamID, requestorID, managerID uuid.UUID) error {
	err := s.teamService.RemoveManager(teamID, requestorID, managerID)
	if err != nil {
		return err
	}
	
	// Cache update is handled by Kafka event handler
	return nil
}

// GetTeam gets team with cached member lookup
func (s *CacheIntegratedTeamService) GetTeam(teamID, userID uuid.UUID) (*models.Team, error) {
	ctx := context.Background()
	
	// Check if user is in team using cache
	cachedMembers, err := s.cacheService.GetTeamMembers(ctx, teamID)
	if err == nil && cachedMembers != nil {
		log.Printf("Cache HIT for team %s members", teamID)
		
		// Check if user is in cached members list
		userInTeam := false
		for _, memberID := range cachedMembers {
			if memberID == userID {
				userInTeam = true
				break
			}
		}
		
		if !userInTeam {
			return nil, fmt.Errorf("access denied: you are not a member of this team")
		}
	}
	
	// Get team from database (could also be cached in the future)
	return s.teamService.GetTeam(teamID, userID)
}

// GetUserTeams gets user teams
func (s *CacheIntegratedTeamService) GetUserTeams(userID uuid.UUID) ([]*models.Team, error) {
	return s.teamService.GetUserTeams(userID)
}

// CacheIntegratedShareService wraps share service with ACL caching
type CacheIntegratedShareService struct {
	shareService ShareService
	cacheService cache.CacheService
}

// NewCacheIntegratedShareService creates a new cache-integrated share service
func NewCacheIntegratedShareService(shareService ShareService, cacheService cache.CacheService) *CacheIntegratedShareService {
	return &CacheIntegratedShareService{
		shareService: shareService,
		cacheService: cacheService,
	}
}

// ShareFolder shares folder and updates ACL cache
func (s *CacheIntegratedShareService) ShareFolder(folderID, ownerID, targetUserID uuid.UUID, accessLevel string) error {
	err := s.shareService.ShareFolder(folderID, ownerID, targetUserID, accessLevel)
	if err != nil {
		return err
	}
	
	// Cache update is handled by Kafka event handler
	return nil
}

// UnshareFolder unshares folder and updates ACL cache
func (s *CacheIntegratedShareService) UnshareFolder(folderID, ownerID, targetUserID uuid.UUID) error {
	err := s.shareService.UnshareFolder(folderID, ownerID, targetUserID)
	if err != nil {
		return err
	}
	
	// Cache update is handled by Kafka event handler
	return nil
}

// GetFolderShares gets folder shares
func (s *CacheIntegratedShareService) GetFolderShares(folderID, userID uuid.UUID) ([]*models.FolderShare, error) {
	return s.shareService.GetFolderShares(folderID, userID)
}

// ShareNote shares note and updates ACL cache
func (s *CacheIntegratedShareService) ShareNote(noteID, ownerID, targetUserID uuid.UUID, accessLevel string) error {
	err := s.shareService.ShareNote(noteID, ownerID, targetUserID, accessLevel)
	if err != nil {
		return err
	}
	
	// Cache update is handled by Kafka event handler
	return nil
}

// UnshareNote unshares note and updates ACL cache
func (s *CacheIntegratedShareService) UnshareNote(noteID, ownerID, targetUserID uuid.UUID) error {
	err := s.shareService.UnshareNote(noteID, ownerID, targetUserID)
	if err != nil {
		return err
	}
	
	// Cache update is handled by Kafka event handler
	return nil
}

// GetNoteShares gets note shares
func (s *CacheIntegratedShareService) GetNoteShares(noteID, userID uuid.UUID) ([]*models.NoteShare, error) {
	return s.shareService.GetNoteShares(noteID, userID)
}

// CheckAssetAccess checks if user has access to asset using cache first
func (s *CacheIntegratedShareService) CheckAssetAccess(assetID, userID uuid.UUID) (string, error) {
	ctx := context.Background()
	
	// Try to get ACL from cache first
	if cachedACL, err := s.cacheService.GetAssetACL(ctx, assetID); err == nil && cachedACL != nil {
		if accessLevel, exists := cachedACL[userID.String()]; exists {
			log.Printf("Cache HIT for asset %s ACL, user %s has %s access", assetID, userID, accessLevel)
			return accessLevel, nil
		}
		log.Printf("Cache HIT for asset %s ACL, but user %s not found", assetID, userID)
		return "", nil
	}
	
	log.Printf("Cache MISS for asset %s ACL", assetID)
	return "", nil // Let the service handle database lookup
}