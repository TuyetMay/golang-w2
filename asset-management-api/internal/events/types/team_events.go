package types

import (
	"time"
	"github.com/google/uuid"
)

// Team event types
const (
	TeamCreated      = "TEAM_CREATED"
	MemberAdded      = "MEMBER_ADDED"
	MemberRemoved    = "MEMBER_REMOVED"
	ManagerAdded     = "MANAGER_ADDED"
	ManagerRemoved   = "MANAGER_REMOVED"
)

// Topics
const (
	TeamActivityTopic = "team.activity"
)

// BaseTeamEvent represents the common fields for all team events
type BaseTeamEvent struct {
	EventType     string    `json:"eventType"`
	TeamID        uuid.UUID `json:"teamId"`
	PerformedBy   uuid.UUID `json:"performedBy"`
	Timestamp     time.Time `json:"timestamp"`
}

// TeamCreatedEvent represents a team creation event
type TeamCreatedEvent struct {
	BaseTeamEvent
	TeamName    string      `json:"teamName"`
	Managers    []uuid.UUID `json:"managers"`
	Members     []uuid.UUID `json:"members"`
}

// MemberChangedEvent represents member addition/removal events
type MemberChangedEvent struct {
	BaseTeamEvent
	TargetUserID uuid.UUID `json:"targetUserId"`
	UserName     string    `json:"userName"`
}

// ManagerChangedEvent represents manager addition/removal events
type ManagerChangedEvent struct {
	BaseTeamEvent
	TargetUserID uuid.UUID `json:"targetUserId"`
	UserName     string    `json:"userName"`
}

// NewTeamCreatedEvent creates a new team creation event
func NewTeamCreatedEvent(teamID, performedBy uuid.UUID, teamName string, managers, members []uuid.UUID) *TeamCreatedEvent {
	return &TeamCreatedEvent{
		BaseTeamEvent: BaseTeamEvent{
			EventType:   TeamCreated,
			TeamID:      teamID,
			PerformedBy: performedBy,
			Timestamp:   time.Now().UTC(),
		},
		TeamName: teamName,
		Managers: managers,
		Members:  members,
	}
}

// NewMemberAddedEvent creates a new member added event
func NewMemberAddedEvent(teamID, performedBy, targetUserID uuid.UUID, userName string) *MemberChangedEvent {
	return &MemberChangedEvent{
		BaseTeamEvent: BaseTeamEvent{
			EventType:   MemberAdded,
			TeamID:      teamID,
			PerformedBy: performedBy,
			Timestamp:   time.Now().UTC(),
		},
		TargetUserID: targetUserID,
		UserName:     userName,
	}
}

// NewMemberRemovedEvent creates a new member removed event
func NewMemberRemovedEvent(teamID, performedBy, targetUserID uuid.UUID, userName string) *MemberChangedEvent {
	return &MemberChangedEvent{
		BaseTeamEvent: BaseTeamEvent{
			EventType:   MemberRemoved,
			TeamID:      teamID,
			PerformedBy: performedBy,
			Timestamp:   time.Now().UTC(),
		},
		TargetUserID: targetUserID,
		UserName:     userName,
	}
}

// NewManagerAddedEvent creates a new manager added event
func NewManagerAddedEvent(teamID, performedBy, targetUserID uuid.UUID, userName string) *ManagerChangedEvent {
	return &ManagerChangedEvent{
		BaseTeamEvent: BaseTeamEvent{
			EventType:   ManagerAdded,
			TeamID:      teamID,
			PerformedBy: performedBy,
			Timestamp:   time.Now().UTC(),
		},
		TargetUserID: targetUserID,
		UserName:     userName,
	}
}

// NewManagerRemovedEvent creates a new manager removed event
func NewManagerRemovedEvent(teamID, performedBy, targetUserID uuid.UUID, userName string) *ManagerChangedEvent {
	return &ManagerChangedEvent{
		BaseTeamEvent: BaseTeamEvent{
			EventType:   ManagerRemoved,
			TeamID:      teamID,
			PerformedBy: performedBy,
			Timestamp:   time.Now().UTC(),
		},
		TargetUserID: targetUserID,
		UserName:     userName,
	}
}