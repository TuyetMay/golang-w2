package service

import (
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	serviceInterfaces "asset-management-api/internal/service/interfaces"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type folderService struct {
	folderRepo interfaces.FolderRepository
	shareRepo  interfaces.ShareRepository
}

func NewFolderService(folderRepo interfaces.FolderRepository, shareRepo interfaces.ShareRepository) serviceInterfaces.FolderService {
	return &folderService{
		folderRepo: folderRepo,
		shareRepo:  shareRepo,
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

	folder, err := s.folderRepo.GetByID(folderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}

	folder.Name = name
	folder.Description = description

	err = s.folderRepo.Update(folder)
	if err != nil {
		return nil, fmt.Errorf("failed to update folder: %w", err)
	}

	return folder, nil
}

func (s *folderService) DeleteFolder(folderID, userID uuid.UUID) error {
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