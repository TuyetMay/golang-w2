package service

import (
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	serviceInterfaces "asset-management-api/internal/service/interfaces"
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type shareService struct {
	shareRepo  interfaces.ShareRepository
	folderRepo interfaces.FolderRepository
	noteRepo   interfaces.NoteRepository
	userRepo   interfaces.UserRepository
}

func NewShareService(shareRepo interfaces.ShareRepository, folderRepo interfaces.FolderRepository, noteRepo interfaces.NoteRepository, userRepo interfaces.UserRepository) serviceInterfaces.ShareService {
	return &shareService{
		shareRepo:  shareRepo,
		folderRepo: folderRepo,
		noteRepo:   noteRepo,
		userRepo:   userRepo,
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
	_, err = s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("target user not found: %w", err)
	}

	// Don't allow sharing with the owner
	if ownerID == targetUserID {
		return errors.New("cannot share folder with yourself")
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

	err = s.shareRepo.UnshareFolder(folderID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to unshare folder: %w", err)
	}

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
	_, err = s.userRepo.GetByID(targetUserID)
	if err != nil {
		return fmt.Errorf("target user not found: %w", err)
	}

	// Don't allow sharing with the owner
	if ownerID == targetUserID {
		return errors.New("cannot share note with yourself")
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

	err = s.shareRepo.UnshareNote(noteID, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to unshare note: %w", err)
	}

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