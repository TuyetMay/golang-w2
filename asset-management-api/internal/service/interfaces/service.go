package interfaces

import (
	"asset-management-api/internal/models"
	"github.com/google/uuid"
)

type FolderService interface {
	CreateFolder(userID uuid.UUID, name, description string) (*models.Folder, error)
	GetFolder(folderID, userID uuid.UUID) (*models.Folder, error)
	UpdateFolder(folderID, userID uuid.UUID, name, description string) (*models.Folder, error)
	DeleteFolder(folderID, userID uuid.UUID) error
	GetUserFolders(userID uuid.UUID) ([]*models.Folder, error)
}

type NoteService interface {
	CreateNote(userID, folderID uuid.UUID, title, body string) (*models.Note, error)
	GetNote(noteID, userID uuid.UUID) (*models.Note, error)
	UpdateNote(noteID, userID uuid.UUID, title, body string) (*models.Note, error)
	DeleteNote(noteID, userID uuid.UUID) error
	GetNotesByFolder(folderID, userID uuid.UUID) ([]*models.Note, error)
	GetUserNotes(userID uuid.UUID) ([]*models.Note, error)
}

type ShareService interface {
	// Folder sharing
	ShareFolder(folderID, ownerID, targetUserID uuid.UUID, accessLevel string) error
	UnshareFolder(folderID, ownerID, targetUserID uuid.UUID) error
	GetFolderShares(folderID, userID uuid.UUID) ([]*models.FolderShare, error)

	// Note sharing
	ShareNote(noteID, ownerID, targetUserID uuid.UUID, accessLevel string) error
	UnshareNote(noteID, ownerID, targetUserID uuid.UUID) error
	GetNoteShares(noteID, userID uuid.UUID) ([]*models.NoteShare, error)
}

type ManagerService interface {
	GetTeamAssets(teamID, managerID uuid.UUID) ([]*models.AssetInfo, error)
	GetUserAssets(targetUserID, managerID uuid.UUID) ([]*models.AssetInfo, error)
}

// Thêm vào cuối file:
type TeamService interface {
	CreateTeam(creatorID uuid.UUID, teamName string, managers []TeamMemberInfo, members []TeamMemberInfo) (*models.Team, error)
	AddMember(teamID, requestorID, memberID uuid.UUID) error
	RemoveMember(teamID, requestorID, memberID uuid.UUID) error  
	AddManager(teamID, requestorID, managerID uuid.UUID) error
	RemoveManager(teamID, requestorID, managerID uuid.UUID) error
	GetTeam(teamID, userID uuid.UUID) (*models.Team, error)
	GetUserTeams(userID uuid.UUID) ([]*models.Team, error)
}

// Và thêm struct:
type TeamMemberInfo struct {
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
}