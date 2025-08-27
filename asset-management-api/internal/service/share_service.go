package service

import (
	"asset-management-api/internal/events/types"
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	serviceInterfaces "asset-management-api/internal/service/interfaces"
	"asset-management-api/pkg/eventbus"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type shareService struct {
	shareRepo  interfaces.ShareRepository
	folderRepo interfaces.FolderRepository
	noteRepo   interfaces.NoteRepository
	userRepo   interfaces.UserRepository
	eventBus   eventbus.EventBus // NEW: Added event bus
}

// NEW: Updated constructor to accept event bus
func NewShareService(shareRepo interfaces.ShareRepository, folderRepo interfaces.FolderRepository, noteRepo interfaces.NoteRepository, userRepo interfaces.UserRepository, eventBus eventbus.EventBus) serviceInterfaces.ShareService {
	return &shareService{
		shareRepo:  shareRepo,
		folderRepo: folderRepo,
		noteRepo:   noteRepo,
		userRepo:   userRepo,
		eventBus:   eventBus,
	}
}

// Folder sharing methods
func (s *shareService) ShareFolder(folderID, ownerID, targetUserID uuid.UUID, accessLevel string) error {
	if accessLevel != "read" && accessLevel != "write" {
		return errors.New("access level must be 'read' or 'write'")
	}

	// Check if the user owns the folder
	isOwner, err := s.folderRepo.CheckOwnership(folderID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to check folder ownership: %w", err)
	}
	if !isOwner {
		return errors.New("access denied: only the folder owner can share it")
	}

	// Check if target user exists
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("target user not found: %w", err)
	}

	// Don't allow sharing with the owner
	if ownerID == targetUserID {
		return errors.New("cannot share folder with yourself")
	}

	// Get owner info for event
	ownerUser, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		return fmt.Errorf("owner user not found: %w", err)
	}

	folderShare := &models.FolderShare{
		FolderID:         folderID,
		SharedWithUserID: targetUserID,
		AccessLevel:      accessLevel,
		SharedBy:         ownerID,
	}

	err = s.shareRepo.ShareFolder(folderShare)
	if err != nil {
		return fmt.Errorf("failed to share folder: %w", err)
	}

	// NEW: Publish folder shared event
	s.publishFolderSharedEvent(folderID, ownerID, targetUserID, accessLevel, ownerUser.Username)

	return nil
}

func (s *shareService) UnshareFolder(folderID, ownerID, targetUserID uuid.UUID) error {
	// Check if the user owns the folder
	isOwner, err := s.folderRepo.CheckOwnership(folderID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to check folder ownership: %w", err)
	}
	if !isOwner {
		return errors.New("access denied: only the folder owner can unshare it")
	}

	// Get owner info for event
	ownerUser, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		return fmt.Errorf("owner user not found: %w", err)
	}

	err = s.shareRepo.UnshareFolder(folderID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to unshare folder: %w", err)
	}

	// NEW: Publish folder unshared event
	s.publishFolderUnsharedEvent(folderID, ownerID, targetUserID, ownerUser.Username)

	return nil
}

func (s *shareService) GetFolderShares(folderID, userID uuid.UUID) ([]*models.FolderShare, error) {
	// Check if the user owns the folder
	isOwner, err := s.folderRepo.CheckOwnership(folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check folder ownership: %w", err)
	}
	if !isOwner {
		return nil, errors.New("access denied: only the folder owner can view shares")
	}

	shares, err := s.shareRepo.GetFolderShares(folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder shares: %w", err)
	}

	return shares, nil
}

// Note sharing methods
func (s *shareService) ShareNote(noteID, ownerID, targetUserID uuid.UUID, accessLevel string) error {
	if accessLevel != "read" && accessLevel != "write" {
		return errors.New("access level must be 'read' or 'write'")
	}

	// Check if the user owns the note
	isOwner, err := s.noteRepo.CheckOwnership(noteID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to check note ownership: %w", err)
	}
	if !isOwner {
		return errors.New("access denied: only the note owner can share it")
	}

	// Check if target user exists
	targetUser, err := s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("target user not found: %w", err)
	}

	// Don't allow sharing with the owner
	if ownerID == targetUserID {
		return errors.New("cannot share note with yourself")
	}

	// Get owner info for event
	ownerUser, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		return fmt.Errorf("owner user not found: %w", err)
	}

	noteShare := &models.NoteShare{
		NoteID:           noteID,
		SharedWithUserID: targetUserID,
		AccessLevel:      accessLevel,
		SharedBy:         ownerID,
	}

	err = s.shareRepo.ShareNote(noteShare)
	if err != nil {
		return fmt.Errorf("failed to share note: %w", err)
	}

	// NEW: Publish note shared event
	s.publishNoteSharedEvent(noteID, ownerID, targetUserID, accessLevel, ownerUser.Username)

	return nil
}

func (s *shareService) UnshareNote(noteID, ownerID, targetUserID uuid.UUID) error {
	// Check if the user owns the note
	isOwner, err := s.noteRepo.CheckOwnership(noteID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to check note ownership: %w", err)
	}
	if !isOwner {
		return errors.New("access denied: only the note owner can unshare it")
	}

	// Get owner info for event
	ownerUser, err := s.userRepo.GetByID(ownerID)
	if err != nil {
		return fmt.Errorf("owner user not found: %w", err)
	}

	err = s.shareRepo.UnshareNote(noteID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to unshare note: %w", err)
	}

	// NEW: Publish note unshared event
	s.publishNoteUnsharedEvent(noteID, ownerID, targetUserID, ownerUser.Username)

	return nil
}

func (s *shareService) GetNoteShares(noteID, userID uuid.UUID) ([]*models.NoteShare, error) {
	// Check if the user owns the note
	isOwner, err := s.noteRepo.CheckOwnership(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check note ownership: %w", err)
	}
	if !isOwner {
		return nil, errors.New("access denied: only the note owner can view shares")
	}

	shares, err := s.shareRepo.GetNoteShares(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note shares: %w", err)
	}

	return shares, nil
}

// NEW: Event publishing methods for folder sharing
func (s *shareService) publishFolderSharedEvent(folderID, ownerID, sharedWithUserID uuid.UUID, accessLevel, sharedByUserName string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewAssetSharedEvent(
		types.FolderShared,
		types.AssetTypeFolder,
		folderID,
		ownerID,
		ownerID, // actionBy is the owner who shared
		sharedWithUserID,
		accessLevel,
		sharedByUserName,
	)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.AssetChangesTopic, event); err != nil {
		log.Printf("Failed to publish folder shared event: %v", err)
	}
}

func (s *shareService) publishFolderUnsharedEvent(folderID, ownerID, unsharedFromUserID uuid.UUID, unsharedByUserName string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewAssetUnsharedEvent(
		types.FolderUnshared,
		types.AssetTypeFolder,
		folderID,
		ownerID,
		ownerID, // actionBy is the owner who unshared
		unsharedFromUserID,
		unsharedByUserName,
	)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.AssetChangesTopic, event); err != nil {
		log.Printf("Failed to publish folder unshared event: %v", err)
	}
}

// NEW: Event publishing methods for note sharing
func (s *shareService) publishNoteSharedEvent(noteID, ownerID, sharedWithUserID uuid.UUID, accessLevel, sharedByUserName string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewAssetSharedEvent(
		types.NoteShared,
		types.AssetTypeNote,
		noteID,
		ownerID,
		ownerID, // actionBy is the owner who shared
		sharedWithUserID,
		accessLevel,
		sharedByUserName,
	)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.AssetChangesTopic, event); err != nil {
		log.Printf("Failed to publish note shared event: %v", err)
	}
}

func (s *shareService) publishNoteUnsharedEvent(noteID, ownerID, unsharedFromUserID uuid.UUID, unsharedByUserName string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewAssetUnsharedEvent(
		types.NoteUnshared,
		types.AssetTypeNote,
		noteID,
		ownerID,
		ownerID, // actionBy is the owner who unshared
		unsharedFromUserID,
		unsharedByUserName,
	)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.AssetChangesTopic, event); err != nil {
		log.Printf("Failed to publish note unshared event: %v", err)
	}
}