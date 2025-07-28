package postgres

import (
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetTeamMembers(teamID uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Table("users").
		Select("users.*").
		Joins("JOIN team_members ON users.user_id = team_members.member_id").
		Where("team_members.team_id = ?", teamID).
		Find(&users).Error
	return users, err
}

func (r *userRepository) CheckIfUserInTeam(userID, teamID uuid.UUID) (bool, error) {
	var count int64
	
	// Check if user is a member
	err := r.db.Model(&models.TeamMember{}).Where("member_id = ? AND team_id = ?", userID, teamID).Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}
	
	// Check if user is a manager
	err = r.db.Model(&models.TeamManager{}).Where("manager_id = ? AND team_id = ?", userID, teamID).Count(&count).Error
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

func (r *userRepository) CheckIfManager(userID uuid.UUID) (bool, error) {
	var user models.User
	err := r.db.First(&user, "user_id = ? AND role = 'manager'", userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
