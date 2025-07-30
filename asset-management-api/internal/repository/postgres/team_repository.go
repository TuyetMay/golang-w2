package postgres

import (
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) interfaces.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(team *models.Team) error {
	return r.db.Create(team).Error
}

func (r *teamRepository) GetByID(teamID uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.Preload("Managers").Preload("Members").First(&team, "team_id = ?", teamID).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamRepository) GetTeamsByManagerID(managerID uuid.UUID) ([]*models.Team, error) {
	var teams []*models.Team
	err := r.db.Table("teams").
		Select("teams.*").
		Joins("JOIN team_managers ON teams.team_id = team_managers.team_id").
		Where("team_managers.manager_id = ?", managerID).
		Preload("Managers").
		Preload("Members").
		Find(&teams).Error
	return teams, err
}

func (r *teamRepository) GetTeamsByMemberID(memberID uuid.UUID) ([]*models.Team, error) {
	var teams []*models.Team
	err := r.db.Table("teams").
		Select("teams.*").
		Joins("JOIN team_members ON teams.team_id = team_members.team_id").
		Where("team_members.member_id = ?", memberID).
		Preload("Managers").
		Preload("Members").
		Find(&teams).Error
	return teams, err
}

func (r *teamRepository) AddManager(teamID, managerID uuid.UUID) error {
	teamManager := &models.TeamManager{
		TeamID:    teamID,
		ManagerID: managerID,
	}
	return r.db.Create(teamManager).Error
}

func (r *teamRepository) RemoveManager(teamID, managerID uuid.UUID) error {
	return r.db.Delete(&models.TeamManager{}, "team_id = ? AND manager_id = ?", teamID, managerID).Error
}

func (r *teamRepository) AddMember(teamID, memberID uuid.UUID) error {
	teamMember := &models.TeamMember{
		TeamID:   teamID,
		MemberID: memberID,
	}
	return r.db.Create(teamMember).Error
}

func (r *teamRepository) RemoveMember(teamID, memberID uuid.UUID) error {
	return r.db.Delete(&models.TeamMember{}, "team_id = ? AND member_id = ?", teamID, memberID).Error
}

func (r *teamRepository) IsTeamManager(teamID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.TeamManager{}).Where("team_id = ? AND manager_id = ?", teamID, userID).Count(&count).Error
	return count > 0, err
}

func (r *teamRepository) IsTeamMember(teamID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.TeamMember{}).Where("team_id = ? AND member_id = ?", teamID, userID).Count(&count).Error
	return count > 0, err
}

func (r *teamRepository) Update(team *models.Team) error {
	return r.db.Save(team).Error
}

func (r *teamRepository) Delete(teamID uuid.UUID) error {
	return r.db.Delete(&models.Team{}, "team_id = ?", teamID).Error
}