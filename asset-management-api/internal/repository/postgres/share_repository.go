package postgres

import (
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type shareRepository struct {
	db *gorm.DB
}

func NewShareRepository(db *gorm.DB) interfaces.ShareRepository {
	return &shareRepository{db: db}
}

// Folder sharing methods
func (r *shareRepository) ShareFolder(folderShare *models.FolderShare) error {
	return r.db.Create(folderShare).Error
}

func (r *shareRepository) UnshareFolder(folderID, userID uuid.UUID) error {
	return r.db.Delete(&models.FolderShare{}, "folder_id = ? AND shared_with_user_id = ?", folderID, userID).Error
}

func (r *shareRepository) GetFolderShares(folderID uuid.UUID) ([]*models.FolderShare, error) {
	var shares []*models.FolderShare
	err := r.db.Preload("SharedWithUser").Preload("SharedByUser").Where("folder_id = ?", folderID).Find(&shares).Error
	return shares, err
}

func (r *shareRepository) CheckFolderAccess(folderID, userID uuid.UUID) (string, error) {
	var share models.FolderShare
	err := r.db.First(&share, "folder_id = ? AND shared_with_user_id = ?", folderID, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return share.AccessLevel, nil
}

// Note sharing methods
func (r *shareRepository) ShareNote(noteShare *models.NoteShare) error {
	return r.db.Create(noteShare).Error
}

func (r *shareRepository) UnshareNote(noteID, userID uuid.UUID) error {
	return r.db.Delete(&models.NoteShare{}, "note_id = ? AND shared_with_user_id = ?", noteID, userID).Error
}

func (r *shareRepository) GetNoteShares(noteID uuid.UUID) ([]*models.NoteShare, error) {
	var shares []*models.NoteShare
	err := r.db.Preload("SharedWithUser").Preload("SharedByUser").Where("note_id = ?", noteID).Find(&shares).Error
	return shares, err
}

func (r *shareRepository) CheckNoteAccess(noteID, userID uuid.UUID) (string, error) {
	var share models.NoteShare
	err := r.db.First(&share, "note_id = ? AND shared_with_user_id = ?", noteID, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return share.AccessLevel, nil
}
