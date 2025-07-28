package postgres

import (
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type folderRepository struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) interfaces.FolderRepository {
	return &folderRepository{db: db}
}

func (r *folderRepository) Create(folder *models.Folder) error {
	return r.db.Create(folder).Error
}

func (r *folderRepository) GetByID(folderID uuid.UUID) (*models.Folder, error) {
	var folder models.Folder
	err := r.db.Preload("Owner").Preload("Notes").First(&folder, "folder_id = ?", folderID).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *folderRepository) GetByOwnerID(ownerID uuid.UUID) ([]*models.Folder, error) {
	var folders []*models.Folder
	err := r.db.Preload("Owner").Where("owner_id = ?", ownerID).Find(&folders).Error
	return folders, err
}

func (r *folderRepository) Update(folder *models.Folder) error {
	return r.db.Save(folder).Error
}

func (r *folderRepository) Delete(folderID uuid.UUID) error {
	// This will cascade delete notes and shares due to foreign key constraints
	return r.db.Delete(&models.Folder{}, "folder_id = ?", folderID).Error
}

func (r *folderRepository) CheckOwnership(folderID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Folder{}).Where("folder_id = ? AND owner_id = ?", folderID, userID).Count(&count).Error
	return count > 0, err
}

func (r *folderRepository) GetSharedFolders(userID uuid.UUID) ([]*models.Folder, error) {
	var folders []*models.Folder
	err := r.db.Table("folders").
		Select("folders.*").
		Joins("JOIN folder_shares ON folders.folder_id = folder_shares.folder_id").
		Where("folder_shares.shared_with_user_id = ?", userID).
		Preload("Owner").
		Find(&folders).Error
	return folders, err
}