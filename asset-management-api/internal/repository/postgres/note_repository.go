package postgres

import (
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type noteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) interfaces.NoteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) Create(note *models.Note) error {
	return r.db.Create(note).Error
}

func (r *noteRepository) GetByID(noteID uuid.UUID) (*models.Note, error) {
	var note models.Note
	err := r.db.Preload("Owner").Preload("Folder").First(&note, "note_id = ?", noteID).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (r *noteRepository) GetByFolderID(folderID uuid.UUID) ([]*models.Note, error) {
	var notes []*models.Note
	err := r.db.Preload("Owner").Where("folder_id = ?", folderID).Find(&notes).Error
	return notes, err
}

func (r *noteRepository) GetByOwnerID(ownerID uuid.UUID) ([]*models.Note, error) {
	var notes []*models.Note
	err := r.db.Preload("Owner").Preload("Folder").Where("owner_id = ?", ownerID).Find(&notes).Error
	return notes, err
}

func (r *noteRepository) Update(note *models.Note) error {
	return r.db.Save(note).Error
}

func (r *noteRepository) Delete(noteID uuid.UUID) error {
	// This will cascade delete shares due to foreign key constraints
	return r.db.Delete(&models.Note{}, "note_id = ?", noteID).Error
}

func (r *noteRepository) CheckOwnership(noteID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Note{}).Where("note_id = ? AND owner_id = ?", noteID, userID).Count(&count).Error
	return count > 0, err
}

func (r *noteRepository) GetSharedNotes(userID uuid.UUID) ([]*models.Note, error) {
	var notes []*models.Note
	err := r.db.Table("notes").
		Select("notes.*").
		Joins("JOIN note_shares ON notes.note_id = note_shares.note_id").
		Where("note_shares.shared_with_user_id = ?", userID).
		Preload("Owner").
		Preload("Folder").
		Find(&notes).Error
	return notes, err
}