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
	"gorm.io/gorm"
)

type folderService struct {
	folderRepo interfaces.FolderRepository
	shareRepo  interfaces.ShareRepository
	eventBus   eventbus.EventBus // NEW: Added event bus
}

// NEW: Updated constructor to accept event bus
func NewFolderService(folderRepo interfaces.FolderRepository, shareRepo interfaces.ShareRepository, eventBus eventbus.EventBus) serviceInterfaces.FolderService {
	return &folderService{
		folderRepo: folderRepo,
		shareRepo:  shareRepo,
		eventBus:   eventBus,
	}
}

func (s *folderService) CreateFolder(userID uuid.UUID, name, description string) (*models.Folder, error) {
	if name == "" {
		return nil, errors.New("folder name is required")
	}

	folder := &models.Folder{
		Name:        name,
		Description: description,
		OwnerID:     userID,
	}

	err := s.folderRepo.Create(folder)
	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	// NEW: Publish folder created event
	s.publishFolderCreatedEvent(folder.FolderID, userID, name, description)

	return folder, nil
}

func (s *folderService) GetFolder(folderID, userID uuid.UUID) (*models.Folder, error) {
	// Check if user owns the folder
	isOwner, err := s.folderRepo.CheckOwnership(folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check folder ownership: %w", err)
	}

	if !isOwner {
		// Check if folder is shared with user
		accessLevel, err := s.shareRepo.CheckFolderAccess(folderID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check folder access: %w", err)
		}
		if accessLevel == "" {
			return nil, errors.New("access denied: you don't have permission to view this folder")
		}
	}

	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}

	return folder, nil
}

func (s *folderService) UpdateFolder(folderID, userID uuid.UUID, name, description string) (*models.Folder, error) {
	if name == "" {
		return nil, errors.New("folder name is required")
	}

	// Get existing folder first
	existingFolder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}

	// Check if user owns the folder or has write access
	isOwner, err := s.folderRepo.CheckOwnership(folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check folder ownership: %w", err)
	}

	if !isOwner {
		accessLevel, err := s.shareRepo.CheckFolderAccess(folderID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check folder access: %w", err)
		}
		if accessLevel != "write" {
			return nil, errors.New("access denied: you don't have write permission for this folder")
		}
	}

	// Track changes for event
	var changes []string
	if existingFolder.Name != name {
		changes = append(changes, "name")
	}
	if existingFolder.Description != description {
		changes = append(changes, "description")
	}

	// Update folder
	existingFolder.Name = name
	existingFolder.Description = description

	err = s.folderRepo.Update(existingFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to update folder: %w", err)
	}

	// NEW: Publish folder updated event if there were changes
	if len(changes) > 0 {
		s.publishFolderUpdatedEvent(folderID, existingFolder.OwnerID, userID, name, description, changes)
	}

	return existingFolder, nil
}

func (s *folderService) DeleteFolder(folderID, userID uuid.UUID) error {
	// Get folder info before deletion
	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("folder not found")
		}
		return fmt.Errorf("failed to get folder: %w", err)
	}

	// Only the owner can delete a folder
	isOwner, err := s.folderRepo.CheckOwnership(folderID, userID)
	if err != nil {
		return fmt.Errorf("failed to check folder ownership: %w", err)
	}

	if !isOwner {
		return errors.New("access denied: only the folder owner can delete it")
	}

	err = s.folderRepo.Delete(folderID)
	if err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}

	// NEW: Publish folder deleted event
	s.publishFolderDeletedEvent(folderID, folder.OwnerID, userID, folder.Name)

	return nil
}

func (s *folderService) GetUserFolders(userID uuid.UUID) ([]*models.Folder, error) {
	// Get owned folders
	ownedFolders, err := s.folderRepo.GetByOwnerID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owned folders: %w", err)
	}

	// Get shared folders
	sharedFolders, err := s.folderRepo.GetSharedFolders(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared folders: %w", err)
	}

	// Combine both lists
	allFolders := append(ownedFolders, sharedFolders...)
	return allFolders, nil
}

// NEW: Event publishing methods
func (s *folderService) publishFolderCreatedEvent(folderID, ownerID uuid.UUID, name, description string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewFolderCreatedEvent(folderID, ownerID, ownerID, name, description)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.AssetChangesTopic, event); err != nil {
		log.Printf("Failed to publish folder created event: %v", err)
	}
}

func (s *folderService) publishFolderUpdatedEvent(folderID, ownerID, actionBy uuid.UUID, name, description string, changes []string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewFolderUpdatedEvent(folderID, ownerID, actionBy, name, description, changes)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.AssetChangesTopic, event); err != nil {
		log.Printf("Failed to publish folder updated event: %v", err)
	}
}

func (s *folderService) publishFolderDeletedEvent(folderID, ownerID, actionBy uuid.UUID, name string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewFolderDeletedEvent(folderID, ownerID, actionBy, name)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.AssetChangesTopic, event); err != nil {
		log.Printf("Failed to publish folder deleted event: %v", err)
	}
}