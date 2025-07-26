package repositories

import (
	"team-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamRepository interface {
	CreateTeam(team *models.Team) error
	GetTeamByID(id uuid.UUID) (*models.Team, error)
	GetTeamsByUser(userID uuid.UUID) ([]models.Team, error)
	UpdateTeam(team *models.Team) error
	DeleteTeam(id uuid.UUID) error
	
	// Member management
	AddMember(teamMember *models.TeamMember) error
	RemoveMember(teamID, userID uuid.UUID) error
	GetTeamMembers(teamID uuid.UUID) ([]models.TeamMember, error)
	IsMember(teamID, userID uuid.UUID) (bool, error)
	
	// Manager management
	AddManager(teamManager *models.TeamManager) error
	RemoveManager(teamID, userID uuid.UUID) error
	GetTeamManagers(teamID uuid.UUID) ([]models.TeamManager, error)
	IsManager(teamID, userID uuid.UUID) (bool, error)
	
	// User operations
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetUsersByIDs(ids []uuid.UUID) ([]models.User, error)
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) CreateTeam(team *models.Team) error {
	return r.db.Create(team).Error
}

func (r *teamRepository) GetTeamByID(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.Preload("Creator").
		Preload("Members.User").
		Preload("Managers.User").
		First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) GetTeamsByUser(userID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team
	
	// Get teams where user is creator, member, or manager
	err := r.db.Preload("Creator").
		Preload("Members.User").
		Preload("Managers.User").
		Where("created_by = ? OR id IN (SELECT team_id FROM team_members WHERE user_id = ?) OR id IN (SELECT team_id FROM team_managers WHERE user_id = ?)", 
			userID, userID, userID).
		Find(&teams).Error
	
	return teams, err
}

func (r *teamRepository) UpdateTeam(team *models.Team) error {
	return r.db.Save(team).Error
}

func (r *teamRepository) DeleteTeam(id uuid.UUID) error {
	return r.db.Delete(&models.Team{}, "id = ?", id).Error
}

// Member management
func (r *teamRepository) AddMember(teamMember *models.TeamMember) error {
	return r.db.Create(teamMember).Error
}

func (r *teamRepository) RemoveMember(teamID, userID uuid.UUID) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&models.TeamMember{}).Error
}

func (r *teamRepository) GetTeamMembers(teamID uuid.UUID) ([]models.TeamMember, error) {
	var members []models.TeamMember
	err := r.db.Preload("User").
		Where("team_id = ?", teamID).
		Find(&members).Error
	return members, err
}

func (r *teamRepository) IsMember(teamID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(&count).Error
	return count > 0, err
}

// Manager management
func (r *teamRepository) AddManager(teamManager *models.TeamManager) error {
	return r.db.Create(teamManager).Error
}

func (r *teamRepository) RemoveManager(teamID, userID uuid.UUID) error {
	return r.db.Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&models.TeamManager{}).Error
}

func (r *teamRepository) GetTeamManagers(teamID uuid.UUID) ([]models.TeamManager, error) {
	var managers []models.TeamManager
	err := r.db.Preload("User").
		Where("team_id = ?", teamID).
		Find(&managers).Error
	return managers, err
}

func (r *teamRepository) IsManager(teamID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.TeamManager{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(&count).Error
	return count > 0, err
}

// User operations
func (r *teamRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *teamRepository) GetUsersByIDs(ids []uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}