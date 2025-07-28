package service

import (
	"asset-management-api/internal/models"
	"asset-management-api/internal/repository/interfaces"
	serviceInterfaces "asset-management-api/internal/service/interfaces"
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type managerService struct {
	userRepo   interfaces.UserRepository
	teamRepo   interfaces.TeamRepository
	folderRepo interfaces.FolderRepository
	noteRepo   interfaces.NoteRepository
	shareRepo  interfaces.ShareRepository
}

func NewManagerService(userRepo interfaces.UserRepository, teamRepo interfaces.TeamRepository, folderRepo interfaces.FolderRepository, noteRepo interfaces.NoteRepository, shareRepo interfaces.ShareRepository) serviceInterfaces.ManagerService {
	return &managerService{
		userRepo:   userRepo,
		teamRepo:   teamRepo,
		folderRepo: folderRepo,
		noteRepo:   noteRepo,
		shareRepo:  shareRepo,
	}
}

func (s *managerService) GetTeamAssets(teamID, managerID uuid.UUID) ([]*models.AssetInfo, error) {
	// Check if user is a manager
	isManager, err := s.userRepo.CheckIfManager(managerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check manager status: %w", err)
	}
	if !isManager {
		return nil, errors.New("access denied: only managers can view team assets")
	}

	// Check if manager belongs to this team
	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	isTeamManager := false
	for _, manager := range team.Managers {
		if manager.UserID == managerID {
			isTeamManager = true
			break
		}
	}
	if !isTeamManager {
		return nil, errors.New("access denied: you are not a manager of this team")
	}

	var allAssets []*models.AssetInfo

	// Get assets for each team member
	for _, member := range team.Members {
		memberAssets, err := s.getUserAssetsInternal(member.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get assets for member %s: %w", member.Username, err)
		}
		allAssets = append(allAssets, memberAssets...)
	}

	return allAssets, nil
}

func (s *managerService) GetUserAssets(targetUserID, managerID uuid.UUID) ([]*models.AssetInfo, error) {
	// Check if user is a manager
	isManager, err := s.userRepo.CheckIfManager(managerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check manager status: %w", err)
	}
	if !isManager {
		return nil, errors.New("access denied: only managers can view user assets")
	}

	// Check if manager and target user are in the same team
	managerTeams, err := s.teamRepo.GetTeamsByManagerID(managerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get manager teams: %w", err)
	}

	userTeams, err := s.teamRepo.GetTeamsByMemberID(targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}

	// Check if they share any team
	shareTeam := false
	for _, managerTeam := range managerTeams {
		for _, userTeam := range userTeams {
			if managerTeam.TeamID == userTeam.TeamID {
				shareTeam = true
				break
			}
		}
		if shareTeam {
			break
		}
	}

	if !shareTeam {
		return nil, errors.New("access denied: you can only view assets of users in your teams")
	}

	return s.getUserAssetsInternal(targetUserID)
}

func (s *managerService) getUserAssetsInternal(userID uuid.UUID) ([]*models.AssetInfo, error) {
	var assets []*models.AssetInfo

	// Get user info
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get owned folders
	folders, err := s.folderRepo.GetByOwnerID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user folders: %w", err)
	}

	for _, folder := range folders {
		assets = append(assets, &models.AssetInfo{
			Type:      "folder",
			ID:        folder.FolderID,
			Name:      folder.Name,
			OwnerID:   folder.OwnerID,
			OwnerName: user.Username,
			CreatedAt: folder.CreatedAt,
		})
	}

	// Get owned notes
	notes, err := s.noteRepo.GetByOwnerID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notes: %w", err)
	}

	for _, note := range notes {
		assets = append(assets, &models.AssetInfo{
			Type:      "note",
			ID:        note.NoteID,
			Name:      note.Title,
			OwnerID:   note.OwnerID,
			OwnerName: user.Username,
			CreatedAt: note.CreatedAt,
		})
	}

	// Get shared folders
	sharedFolders, err := s.folderRepo.GetSharedFolders(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared folders: %w", err)
	}

	for _, folder := range sharedFolders {
		accessLevel, _ := s.shareRepo.CheckFolderAccess(folder.FolderID, userID)
		assets = append(assets, &models.AssetInfo{
			Type:        "folder",
			ID:          folder.FolderID,
			Name:        folder.Name,
			OwnerID:     folder.OwnerID,
			OwnerName:   folder.Owner.Username,
			AccessLevel: accessLevel,
			CreatedAt:   folder.CreatedAt,
		})
	}

	// Get shared notes
	sharedNotes, err := s.noteRepo.GetSharedNotes(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared notes: %w", err)
	}

	for _, note := range sharedNotes {
		accessLevel, _ := s.shareRepo.CheckNoteAccess(note.NoteID, userID)
		assets = append(assets, &models.AssetInfo{
			Type:        "note",
			ID:          note.NoteID,
			Name:        note.Title,
			OwnerID:     note.OwnerID,
			OwnerName:   note.Owner.Username,
			AccessLevel: accessLevel,
			CreatedAt:   note.CreatedAt,
		})
	}

	return assets, nil
}