package service

import (
	"asset-management-api/internal/events/types"
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	serviceInterfaces "asset-management-api/internal/service/interfaces"
	"asset-management-api/pkg/eventbus"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
)

type teamService struct {
	teamRepo  interfaces.TeamRepository
	userRepo  interfaces.UserRepository
	eventBus  eventbus.EventBus // NEW: Added event bus
}

// NEW: Updated constructor to accept event bus
func NewTeamService(teamRepo interfaces.TeamRepository, userRepo interfaces.UserRepository, eventBus eventbus.EventBus) serviceInterfaces.TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		eventBus: eventBus,
	}
}

func (s *teamService) CreateTeam(creatorID uuid.UUID, teamName string, managers []serviceInterfaces.TeamMemberInfo, members []serviceInterfaces.TeamMemberInfo) (*models.Team, error) {
	if teamName == "" {
		return nil, errors.New("team name is required")
	}

	// Check if creator is a manager
	isManager, err := s.userRepo.CheckIfManager(creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check creator role: %w", err)
	}
	if !isManager {
		return nil, errors.New("access denied: only managers can create teams")
	}

	// Create team
	team := &models.Team{
		TeamName:  teamName,
		CreatedBy: creatorID,
	}

	err = s.teamRepo.Create(team)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Add creator as manager
	err = s.teamRepo.AddManager(team.TeamID, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to add creator as manager: %w", err)
	}

	var managerIDs []uuid.UUID
	var memberIDs []uuid.UUID

	// Add additional managers
	for _, manager := range managers {
		managerID, err := uuid.Parse(manager.UserID)
		if err != nil {
			continue // Skip invalid UUIDs
		}
		
		// Check if user exists and has manager role
		user, err := s.userRepo.GetByID(managerID)
		if err != nil {
			continue // Skip non-existent users
		}
		if user.Role != "manager" {
			continue // Skip non-managers
		}

		// Don't add creator again
		if managerID != creatorID {
			s.teamRepo.AddManager(team.TeamID, managerID)
			managerIDs = append(managerIDs, managerID)
		}
	}

	// Add creator to manager list for event
	managerIDs = append(managerIDs, creatorID)

	// Add members
	for _, member := range members {
		memberID, err := uuid.Parse(member.UserID)
		if err != nil {
			continue // Skip invalid UUIDs
		}
		
		// Check if user exists
		_, err = s.userRepo.GetByID(memberID)
		if err != nil {
			continue // Skip non-existent users
		}

		s.teamRepo.AddMember(team.TeamID, memberID)
		memberIDs = append(memberIDs, memberID)
	}

	// NEW: Publish team created event
	s.publishTeamCreatedEvent(team.TeamID, creatorID, teamName, managerIDs, memberIDs)

	// Get the complete team with relationships
	return s.teamRepo.GetByID(team.TeamID)
}

func (s *teamService) AddMember(teamID, requestorID, memberID uuid.UUID) error {
	// Check if requestor is a manager of the team
	isTeamManager, err := s.teamRepo.IsTeamManager(teamID, requestorID)
	if err != nil {
		return fmt.Errorf("failed to check team manager status: %w", err)
	}
	if !isTeamManager {
		return errors.New("access denied: only team managers can add members")
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(memberID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user is already a team member or manager
	isAlreadyMember, err := s.userRepo.CheckIfUserInTeam(memberID, teamID)
	if err != nil {
		return fmt.Errorf("failed to check team membership: %w", err)
	}
	if isAlreadyMember {
		return errors.New("user is already a member of this team")
	}

	err = s.teamRepo.AddMember(teamID, memberID)
	if err != nil {
		return err
	}

	// NEW: Publish member added event
	s.publishMemberAddedEvent(teamID, requestorID, memberID, user.Username)

	return nil
}

func (s *teamService) RemoveMember(teamID, requestorID, memberID uuid.UUID) error {
	// Check if requestor is a manager of the team
	isTeamManager, err := s.teamRepo.IsTeamManager(teamID, requestorID)
	if err != nil {
		return fmt.Errorf("failed to check team manager status: %w", err)
	}
	if !isTeamManager {
		return errors.New("access denied: only team managers can remove members")
	}

	// Get user info before removal
	user, err := s.userRepo.GetByID(memberID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if member exists in team
	isMember, err := s.teamRepo.IsTeamMember(teamID, memberID)
	if err != nil {
		return fmt.Errorf("failed to check team membership: %w", err)
	}
	if !isMember {
		return errors.New("member not found in team")
	}

	err = s.teamRepo.RemoveMember(teamID, memberID)
	if err != nil {
		return err
	}

	// NEW: Publish member removed event
	s.publishMemberRemovedEvent(teamID, requestorID, memberID, user.Username)

	return nil
}

func (s *teamService) AddManager(teamID, requestorID, managerID uuid.UUID) error {
	// Check if requestor is a manager of the team
	isTeamManager, err := s.teamRepo.IsTeamManager(teamID, requestorID)
	if err != nil {
		return fmt.Errorf("failed to check team manager status: %w", err)
	}
	if !isTeamManager {
		return errors.New("access denied: only team managers can add other managers")
	}

	// Check if target user exists and has manager role
	user, err := s.userRepo.GetByID(managerID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if user.Role != "manager" {
		return errors.New("target user must have manager role")
	}

	// Check if user is already a manager
	isAlreadyManager, err := s.teamRepo.IsTeamManager(teamID, managerID)
	if err != nil {
		return fmt.Errorf("failed to check manager status: %w", err)
	}
	if isAlreadyManager {
		return errors.New("user is already a manager of this team")
	}

	// Remove from members if they are a member
	isMember, _ := s.teamRepo.IsTeamMember(teamID, managerID)
	if isMember {
		s.teamRepo.RemoveMember(teamID, managerID)
	}

	err = s.teamRepo.AddManager(teamID, managerID)
	if err != nil {
		return err
	}

	// NEW: Publish manager added event
	s.publishManagerAddedEvent(teamID, requestorID, managerID, user.Username)

	return nil
}

func (s *teamService) RemoveManager(teamID, requestorID, managerID uuid.UUID) error {
	// Check if requestor is a manager of the team
	isTeamManager, err := s.teamRepo.IsTeamManager(teamID, requestorID)
	if err != nil {
		return fmt.Errorf("failed to check team manager status: %w", err)
	}
	if !isTeamManager {
		return errors.New("access denied: only team managers can remove other managers")
	}

	// Get team to check creator
	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	// Cannot remove the team creator
	if team.CreatedBy == managerID {
		return errors.New("cannot remove the team creator")
	}

	// Get user info before removal
	user, err := s.userRepo.GetByID(managerID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if target is actually a manager
	isManager, err := s.teamRepo.IsTeamManager(teamID, managerID)
	if err != nil {
		return fmt.Errorf("failed to check manager status: %w", err)
	}
	if !isManager {
		return errors.New("manager not found in team")
	}

	err = s.teamRepo.RemoveManager(teamID, managerID)
	if err != nil {
		return err
	}

	// NEW: Publish manager removed event
	s.publishManagerRemovedEvent(teamID, requestorID, managerID, user.Username)

	return nil
}

func (s *teamService) GetTeam(teamID, userID uuid.UUID) (*models.Team, error) {
	// Check if user is part of the team (as member or manager)
	isInTeam, err := s.userRepo.CheckIfUserInTeam(userID, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to check team membership: %w", err)
	}
	if !isInTeam {
		return nil, errors.New("access denied: you are not a member of this team")
	}

	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("team not found")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return team, nil
}

func (s *teamService) GetUserTeams(userID uuid.UUID) ([]*models.Team, error) {
	// Get teams where user is a manager
	managerTeams, err := s.teamRepo.GetTeamsByManagerID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get manager teams: %w", err)
	}

	// Get teams where user is a member
	memberTeams, err := s.teamRepo.GetTeamsByMemberID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member teams: %w", err)
	}

	// Combine teams and remove duplicates
	teamMap := make(map[uuid.UUID]*models.Team)
	
	for _, team := range managerTeams {
		teamMap[team.TeamID] = team
	}
	
	for _, team := range memberTeams {
		teamMap[team.TeamID] = team
	}

	var allTeams []*models.Team
	for _, team := range teamMap {
		allTeams = append(allTeams, team)
	}

	return allTeams, nil
}

// NEW: Event publishing methods
func (s *teamService) publishTeamCreatedEvent(teamID, performedBy uuid.UUID, teamName string, managers, members []uuid.UUID) {
	if s.eventBus == nil {
		return
	}

	event := types.NewTeamCreatedEvent(teamID, performedBy, teamName, managers, members)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.TeamActivityTopic, event); err != nil {
		log.Printf("Failed to publish team created event: %v", err)
	}
}

func (s *teamService) publishMemberAddedEvent(teamID, performedBy, targetUserID uuid.UUID, userName string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewMemberAddedEvent(teamID, performedBy, targetUserID, userName)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.TeamActivityTopic, event); err != nil {
		log.Printf("Failed to publish member added event: %v", err)
	}
}

func (s *teamService) publishMemberRemovedEvent(teamID, performedBy, targetUserID uuid.UUID, userName string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewMemberRemovedEvent(teamID, performedBy, targetUserID, userName)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.TeamActivityTopic, event); err != nil {
		log.Printf("Failed to publish member removed event: %v", err)
	}
}

func (s *teamService) publishManagerAddedEvent(teamID, performedBy, targetUserID uuid.UUID, userName string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewManagerAddedEvent(teamID, performedBy, targetUserID, userName)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.TeamActivityTopic, event); err != nil {
		log.Printf("Failed to publish manager added event: %v", err)
	}
}

func (s *teamService) publishManagerRemovedEvent(teamID, performedBy, targetUserID uuid.UUID, userName string) {
	if s.eventBus == nil {
		return
	}

	event := types.NewManagerRemovedEvent(teamID, performedBy, targetUserID, userName)
	
	ctx := context.Background()
	if err := s.eventBus.Publish(ctx, types.TeamActivityTopic, event); err != nil {
		log.Printf("Failed to publish manager removed event: %v", err)
	}
}