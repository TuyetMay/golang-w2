package services

import (
	"errors"
	"time"
	"team-service/internal/models"
	"team-service/internal/repositories"
	"github.com/google/uuid"
)

type TeamService interface {
	CreateTeam(userID uuid.UUID, req *models.CreateTeamRequest) (*models.TeamResponse, error)
	GetTeam(teamID uuid.UUID, userID uuid.UUID) (*models.TeamResponse, error)
	GetUserTeams(userID uuid.UUID) ([]models.TeamResponse, error)
	
	AddMember(teamID, userID, memberID uuid.UUID) error
	RemoveMember(teamID, userID, memberID uuid.UUID) error
	
	AddManager(teamID, userID, managerID uuid.UUID) error
	RemoveManager(teamID, userID, managerID uuid.UUID) error
}

type teamService struct {
	repo repositories.TeamRepository
}

func NewTeamService(repo repositories.TeamRepository) TeamService {
	return &teamService{repo: repo}
}

func (s *teamService) CreateTeam(userID uuid.UUID, req *models.CreateTeamRequest) (*models.TeamResponse, error) {
	// Verify user exists and is a manager
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	
	if user.Role != "manager" {
		return nil, errors.New("only managers can create teams")
	}
	
	// Create team
	team := &models.Team{
		Name:      req.TeamName,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := s.repo.CreateTeam(team); err != nil {
		return nil, err
	}
	
	// Add creator as manager
	creatorManager := &models.TeamManager{
		TeamID:  team.ID,
		UserID:  userID,
		AddedBy: userID,
		AddedAt: time.Now(),
	}
	
	if err := s.repo.AddManager(creatorManager); err != nil {
		return nil, err
	}
	
	// Add additional managers
	for _, managerInfo := range req.Managers {
		managerID, err := uuid.Parse(managerInfo.ID)
		if err != nil {
			continue
		}
		
		// Verify user exists and is a manager
		managerUser, err := s.repo.GetUserByID(managerID)
		if err != nil || managerUser.Role != "manager" {
			continue
		}
		
		teamManager := &models.TeamManager{
			TeamID:  team.ID,
			UserID:  managerID,
			AddedBy: userID,
			AddedAt: time.Now(),
		}
		s.repo.AddManager(teamManager)
	}
	
	// Add members
	for _, memberInfo := range req.Members {
		memberID, err := uuid.Parse(memberInfo.ID)
		if err != nil {
			continue
		}
		
		// Verify user exists
		_, err = s.repo.GetUserByID(memberID)
		if err != nil {
			continue
		}
		
		teamMember := &models.TeamMember{
			TeamID:  team.ID,
			UserID:  memberID,
			AddedBy: userID,
			AddedAt: time.Now(),
		}
		s.repo.AddMember(teamMember)
	}
	
	// Return team response
	return s.buildTeamResponse(team.ID)
}

func (s *teamService) GetTeam(teamID uuid.UUID, userID uuid.UUID) (*models.TeamResponse, error) {
	// Check if user has access to team
	if !s.hasTeamAccess(teamID, userID) {
		return nil, errors.New("access denied")
	}
	
	return s.buildTeamResponse(teamID)
}

func (s *teamService) GetUserTeams(userID uuid.UUID) ([]models.TeamResponse, error) {
	teams, err := s.repo.GetTeamsByUser(userID)
	if err != nil {
		return nil, err
	}
	
	var responses []models.TeamResponse
	for _, team := range teams {
		response, err := s.buildTeamResponseFromModel(&team)
		if err != nil {
			continue
		}
		responses = append(responses, *response)
	}
	
	return responses, nil
}

func (s *teamService) AddMember(teamID, userID, memberID uuid.UUID) error {
	// Check if user is manager of the team
	if !s.isTeamManager(teamID, userID) {
		return errors.New("only team managers can add members")
	}
	
	// Verify member exists
	_, err := s.repo.GetUserByID(memberID)
	if err != nil {
		return errors.New("member not found")
	}
	
	// Check if already a member
	isMember, _ := s.repo.IsMember(teamID, memberID)
	if isMember {
		return errors.New("user is already a member")
	}
	
	teamMember := &models.TeamMember{
		TeamID:  teamID,
		UserID:  memberID,
		AddedBy: userID,
		AddedAt: time.Now(),
	}
	
	return s.repo.AddMember(teamMember)
}

func (s *teamService) RemoveMember(teamID, userID, memberID uuid.UUID) error {
	// Check if user is manager of the team
	if !s.isTeamManager(teamID, userID) {
		return errors.New("only team managers can remove members")
	}
	
	// Check if user is actually a member
	isMember, _ := s.repo.IsMember(teamID, memberID)
	if !isMember {
		return errors.New("user is not a member of this team")
	}
	
	return s.repo.RemoveMember(teamID, memberID)
}

func (s *teamService) AddManager(teamID, userID, managerID uuid.UUID) error {
	// Check if user is manager of the team
	if !s.isTeamManager(teamID, userID) {
		return errors.New("only team managers can add other managers")
	}
	
	// Verify manager exists and has manager role
	manager, err := s.repo.GetUserByID(managerID)
	if err != nil {
		return errors.New("manager not found")
	}
	
	if manager.Role != "manager" {
		return errors.New("user must have manager role")
	}
	
	// Check if already a manager
	isManager, _ := s.repo.IsManager(teamID, managerID)
	if isManager {
		return errors.New("user is already a manager")
	}
	
	teamManager := &models.TeamManager{
		TeamID:  teamID,
		UserID:  managerID,
		AddedBy: userID,
		AddedAt: time.Now(),
	}
	
	return s.repo.AddManager(teamManager)
}

func (s *teamService) RemoveManager(teamID, userID, managerID uuid.UUID) error {
	// Get team info
	team, err := s.repo.GetTeamByID(teamID)
	if err != nil {
		return errors.New("team not found")
	}
	
	// Check if user is the creator or a manager
	if team.CreatedBy != userID && !s.isTeamManager(teamID, userID) {
		return errors.New("only team creator or managers can remove managers")
	}
	
	// Prevent removing the team creator
	if team.CreatedBy == managerID {
		return errors.New("cannot remove team creator")
	}
	
	// Check if user is actually a manager
	isManager, _ := s.repo.IsManager(teamID, managerID)
	if !isManager {
		return errors.New("user is not a manager of this team")
	}
	
	return s.repo.RemoveManager(teamID, managerID)
}

// Helper functions
func (s *teamService) hasTeamAccess(teamID, userID uuid.UUID) bool {
	team, err := s.repo.GetTeamByID(teamID)
	if err != nil {
		return false
	}
	
	// Creator has access
	if team.CreatedBy == userID {
		return true
	}
	
	// Check if manager
	isManager, _ := s.repo.IsManager(teamID, userID)
	if isManager {
		return true
	}
	
	// Check if member
	isMember, _ := s.repo.IsMember(teamID, userID)
	return isMember
}

func (s *teamService) isTeamManager(teamID, userID uuid.UUID) bool {
	team, err := s.repo.GetTeamByID(teamID)
	if err != nil {
		return false
	}
	
	// Creator is always a manager
	if team.CreatedBy == userID {
		return true
	}
	
	// Check if explicitly a manager
	isManager, _ := s.repo.IsManager(teamID, userID)
	return isManager
}

func (s *teamService) buildTeamResponse(teamID uuid.UUID) (*models.TeamResponse, error) {
	team, err := s.repo.GetTeamByID(teamID)
	if err != nil {
		return nil, err
	}
	
	return s.buildTeamResponseFromModel(team)
}

func (s *teamService) buildTeamResponseFromModel(team *models.Team) (*models.TeamResponse, error) {
	response := &models.TeamResponse{
		ID:        team.ID,
		Name:      team.Name,
		CreatedBy: team.CreatedBy,
		CreatedAt: team.CreatedAt,
		Members:   []models.UserInfo{},
		Managers:  []models.UserInfo{},
	}
	
	// Add members
	for _, member := range team.Members {
		response.Members = append(response.Members, models.UserInfo{
			ID:       member.User.ID,
			Username: member.User.Username,
			Email:    member.User.Email,
			Role:     member.User.Role,
		})
	}
	
	// Add managers
	for _, manager := range team.Managers {
		response.Managers = append(response.Managers, models.UserInfo{
			ID:       manager.User.ID,
			Username: manager.User.Username,
			Email:    manager.User.Email,
			Role:     manager.User.Role,
		})
	}
	
	return response, nil
}