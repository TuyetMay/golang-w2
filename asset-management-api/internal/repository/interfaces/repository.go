package interfaces

import (
	"asset-management-api/internal/models"
	"github.com/google/uuid"
)

type FolderRepository interface {
	Create(folder *models.Folder) error
	GetByID(folderID uuid.UUID) (*models.Folder, error)
	GetByOwnerID(ownerID uuid.UUID) ([]*models.Folder, error)
	Update(folder *models.Folder) error
	Delete(folderID uuid.UUID) error
	CheckOwnership(folderID, userID uuid.UUID) (bool, error)
	GetSharedFolders(userID uuid.UUID) ([]*models.Folder, error)
}

type NoteRepository interface {
	Create(note *models.Note) error
	GetByID(noteID uuid.UUID) (*models.Note, error)
	GetByFolderID(folderID uuid.UUID) ([]*models.Note, error)
	GetByOwnerID(ownerID uuid.UUID) ([]*models.Note, error)
	Update(note *models.Note) error
	Delete(noteID uuid.UUID) error
	CheckOwnership(noteID, userID uuid.UUID) (bool, error)
	GetSharedNotes(userID uuid.UUID) ([]*models.Note, error)
}

type ShareRepository interface {
	// Folder sharing
	ShareFolder(folderShare *models.FolderShare) error
	UnshareFolder(folderID, userID uuid.UUID) error
	GetFolderShares(folderID uuid.UUID) ([]*models.FolderShare, error)
	CheckFolderAccess(folderID, userID uuid.UUID) (string, error) // returns access level or empty

	// Note sharing
	ShareNote(noteShare *models.NoteShare) error
	UnshareNote(noteID, userID uuid.UUID) error
	GetNoteShares(noteID uuid.UUID) ([]*models.NoteShare, error)
	CheckNoteAccess(noteID, userID uuid.UUID) (string, error) // returns access level or empty
}

type UserRepository interface {
	GetByID(userID uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetTeamMembers(teamID uuid.UUID) ([]*models.User, error)
	CheckIfUserInTeam(userID, teamID uuid.UUID) (bool, error)
	CheckIfManager(userID uuid.UUID) (bool, error)
}

type TeamRepository interface {
	Create(team *models.Team) error
	GetByID(teamID uuid.UUID) (*models.Team, error)
	GetTeamsByManagerID(managerID uuid.UUID) ([]*models.Team, error)
	GetTeamsByMemberID(memberID uuid.UUID) ([]*models.Team, error)
	AddManager(teamID, managerID uuid.UUID) error
	RemoveManager(teamID, managerID uuid.UUID) error
	AddMember(teamID, memberID uuid.UUID) error
	RemoveMember(teamID, memberID uuid.UUID) error
	IsTeamManager(teamID, userID uuid.UUID) (bool, error)
	IsTeamMember(teamID, userID uuid.UUID) (bool, error)
	Update(team *models.Team) error
	Delete(teamID uuid.UUID) error
}