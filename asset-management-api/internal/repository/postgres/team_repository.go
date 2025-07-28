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