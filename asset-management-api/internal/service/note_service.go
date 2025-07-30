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

type noteService struct {
	noteRepo   interfaces.NoteRepository
	folderRepo interfaces.FolderRepository
	shareRepo  interfaces.ShareRepository
}

func NewNoteService(noteRepo interfaces.NoteRepository, folderRepo interfaces.FolderRepository, shareRepo interfaces.ShareRepository) serviceInterfaces.NoteService {
	return &noteService{
		noteRepo:   noteRepo,
		folderRepo: folderRepo,
		shareRepo:  shareRepo,
	}
}

func (s *noteService) CreateNote(userID, folderID uuid.UUID, title, body string) (*models.Note, error) {
	if title == "" {
		return nil, errors.New("note title is required")
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

	note := &models.Note{
		Title:    title,
		Body:     body,
		FolderID: folderID,
		OwnerID:  userID,
	}

	err = s.noteRepo.Create(note)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

func (s *noteService) GetNote(noteID, userID uuid.UUID) (*models.Note, error) {
	// Check if user owns the note
	isOwner, err := s.noteRepo.CheckOwnership(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check note ownership: %w", err)
	}

	if !isOwner {
		// Check if note is shared with user
		accessLevel, err := s.shareRepo.CheckNoteAccess(noteID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check note access: %w", err)
		}
		if accessLevel == "" {
			// Check if user has access to the folder containing this note
			note, err := s.noteRepo.GetByID(noteID)
			if err != nil {
				return nil, fmt.Errorf("failed to get note: %w", err)
			}
			folderAccessLevel, err := s.shareRepo.CheckFolderAccess(note.FolderID, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to check folder access: %w", err)
			}
			if folderAccessLevel == "" {
				return nil, errors.New("access denied: you don't have permission to view this note")
			}
		}
	}

	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("note not found")
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}

func (s *noteService) UpdateNote(noteID, userID uuid.UUID, title, body string) (*models.Note, error) {
	if title == "" {
		return nil, errors.New("note title is required")
	}

	// Check if user owns the note or has write access
	isOwner, err := s.noteRepo.CheckOwnership(noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check note ownership: %w", err)
	}

	if !isOwner {
		accessLevel, err := s.shareRepo.CheckNoteAccess(noteID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check note access: %w", err)
		}
		if accessLevel != "write" {
			// Check folder access as fallback
			note, err := s.noteRepo.GetByID(noteID)
			if err != nil {
				return nil, fmt.Errorf("failed to get note: %w", err)
			}
			folderAccessLevel, err := s.shareRepo.CheckFolderAccess(note.FolderID, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to check folder access: %w", err)
			}
			if folderAccessLevel != "write" {
				return nil, errors.New("access denied: you don't have write permission for this note")
			}
		}
	}

	note, err := s.noteRepo.GetByID(noteID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("note not found")
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	note.Title = title
	note.Body = body

	err = s.noteRepo.Update(note)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return note, nil
}

func (s *noteService) DeleteNote(noteID, userID uuid.UUID) error {
	// Only the owner can delete a note
	isOwner, err := s.noteRepo.CheckOwnership(noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to check note ownership: %w", err)
	}

	if !isOwner {
		return errors.New("access denied: only the note owner can delete it")
	}

	err = s.noteRepo.Delete(noteID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

func (s *noteService) GetNotesByFolder(folderID, userID uuid.UUID) ([]*models.Note, error) {
	// Check if user has access to the folder
	isOwner, err := s.folderRepo.CheckOwnership(folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check folder ownership: %w", err)
	}

	if !isOwner {
		accessLevel, err := s.shareRepo.CheckFolderAccess(folderID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check folder access: %w", err)
		}
		if accessLevel == "" {
			return nil, errors.New("access denied: you don't have permission to view this folder")
		}
	}

	notes, err := s.noteRepo.GetByFolderID(folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}

	return notes, nil
}

func (s *noteService) GetUserNotes(userID uuid.UUID) ([]*models.Note, error) {
	// Get owned notes
	ownedNotes, err := s.noteRepo.GetByOwnerID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owned notes: %w", err)
	}

	// Get shared notes
	sharedNotes, err := s.noteRepo.GetSharedNotes(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared notes: %w", err)
	}

	// Combine both lists
	allNotes := append(ownedNotes, sharedNotes...)
	return allNotes, nil
}