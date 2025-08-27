package handlers

import (
	"asset-management-api/internal/events/types"
	"context"
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
)

// TeamEventHandler handles team-related events
type TeamEventHandler struct {
	db *gorm.DB
}

// NewTeamEventHandler creates a new team event handler
func NewTeamEventHandler(db *gorm.DB) *TeamEventHandler {
	return &TeamEventHandler{db: db}
}

// HandleTeamEvent processes team events
func (h *TeamEventHandler) HandleTeamEvent(ctx context.Context, eventData []byte) error {
	// Parse the base event to determine the event type
	var baseEvent types.BaseTeamEvent
	if err := json.Unmarshal(eventData, &baseEvent); err != nil {
		log.Printf("Failed to parse team event: %v", err)
		return err
	}

	log.Printf("Processing team event: %s for team %s", baseEvent.EventType, baseEvent.TeamID)

	switch baseEvent.EventType {
	case types.TeamCreated:
		return h.handleTeamCreated(ctx, eventData)
	case types.MemberAdded:
		return h.handleMemberAdded(ctx, eventData)
	case types.MemberRemoved:
		return h.handleMemberRemoved(ctx, eventData)
	case types.ManagerAdded:
		return h.handleManagerAdded(ctx, eventData)
	case types.ManagerRemoved:
		return h.handleManagerRemoved(ctx, eventData)
	default:
		log.Printf("Unknown team event type: %s", baseEvent.EventType)
		return nil
	}
}

// handleTeamCreated processes team creation events
func (h *TeamEventHandler) handleTeamCreated(ctx context.Context, eventData []byte) error {
	var event types.TeamCreatedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	// Log the team creation for audit purposes
	auditLog := TeamAuditLog{
		TeamID:      event.TeamID,
		EventType:   event.EventType,
		PerformedBy: event.PerformedBy,
		Details: map[string]interface{}{
			"team_name":      event.TeamName,
			"managers_count": len(event.Managers),
			"members_count":  len(event.Members),
		},
		Timestamp: event.Timestamp,
	}

	return h.saveAuditLog(ctx, auditLog)
}

// handleMemberAdded processes member addition events
func (h *TeamEventHandler) handleMemberAdded(ctx context.Context, eventData []byte) error {
	var event types.MemberChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	// Log the member addition
	auditLog := TeamAuditLog{
		TeamID:      event.TeamID,
		EventType:   event.EventType,
		PerformedBy: event.PerformedBy,
		Details: map[string]interface{}{
			"target_user_id": event.TargetUserID,
			"user_name":      event.UserName,
		},
		Timestamp: event.Timestamp,
	}

	// Send notification (example)
	go h.sendNotification(ctx, NotificationRequest{
		Type:      "team_member_added",
		TeamID:    event.TeamID,
		UserID:    event.TargetUserID,
		Message:   fmt.Sprintf("%s has been added to the team", event.UserName),
		Timestamp: event.Timestamp,
	})

	return h.saveAuditLog(ctx, auditLog)
}

// handleMemberRemoved processes member removal events
func (h *TeamEventHandler) handleMemberRemoved(ctx context.Context, eventData []byte) error {
	var event types.MemberChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	// Log the member removal
	auditLog := TeamAuditLog{
		TeamID:      event.TeamID,
		EventType:   event.EventType,
		PerformedBy: event.PerformedBy,
		Details: map[string]interface{}{
			"target_user_id": event.TargetUserID,
			"user_name":      event.UserName,
		},
		Timestamp: event.Timestamp,
	}

	// Send notification
	go h.sendNotification(ctx, NotificationRequest{
		Type:      "team_member_removed",
		TeamID:    event.TeamID,
		UserID:    event.TargetUserID,
		Message:   fmt.Sprintf("%s has been removed from the team", event.UserName),
		Timestamp: event.Timestamp,
	})

	return h.saveAuditLog(ctx, auditLog)
}

// handleManagerAdded processes manager addition events
func (h *TeamEventHandler) handleManagerAdded(ctx context.Context, eventData []byte) error {
	var event types.ManagerChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	// Log the manager addition
	auditLog := TeamAuditLog{
		TeamID:      event.TeamID,
		EventType:   event.EventType,
		PerformedBy: event.PerformedBy,
		Details: map[string]interface{}{
			"target_user_id": event.TargetUserID,
			"user_name":      event.UserName,
		},
		Timestamp: event.Timestamp,
	}

	// Send notification
	go h.sendNotification(ctx, NotificationRequest{
		Type:      "team_manager_added",
		TeamID:    event.TeamID,
		UserID:    event.TargetUserID,
		Message:   fmt.Sprintf("%s has been promoted to team manager", event.UserName),
		Timestamp: event.Timestamp,
	})

	return h.saveAuditLog(ctx, auditLog)
}

// handleManagerRemoved processes manager removal events
func (h *TeamEventHandler) handleManagerRemoved(ctx context.Context, eventData []byte) error {
	var event types.ManagerChangedEvent
	if err := json.Unmarshal(eventData, &event); err != nil {
		return err
	}

	// Log the manager removal
	auditLog := TeamAuditLog{
		TeamID:      event.TeamID,
		EventType:   event.EventType,
		PerformedBy: event.PerformedBy,
		Details: map[string]interface{}{
			"target_user_id": event.TargetUserID,
			"user_name":      event.UserName,
		},
		Timestamp: event.Timestamp,
	}

	return h.saveAuditLog(ctx, auditLog)
}

// saveAuditLog saves audit log to database
func (h *TeamEventHandler) saveAuditLog(ctx context.Context, auditLog TeamAuditLog) error {
	result := h.db.WithContext(ctx).Create(&auditLog)
	if result.Error != nil {
		log.Printf("Failed to save team audit log: %v", result.Error)
		return result.Error
	}
	
	log.Printf("Team audit log saved: %s for team %s", auditLog.EventType, auditLog.TeamID)
	return nil
}

// sendNotification sends notifications (example implementation)
func (h *TeamEventHandler) sendNotification(ctx context.Context, req NotificationRequest) {
	// This is a placeholder - in a real implementation, you might:
	// 1. Send email notifications
	// 2. Send push notifications
	// 3. Update a notification service
	// 4. Send webhooks to external systems
	
	log.Printf("Sending notification: %s to user %s for team %s", 
		req.Message, req.UserID, req.TeamID)
	
	// Example: Save notification to database
	notification := Notification{
		Type:      req.Type,
		TeamID:    req.TeamID,
		UserID:    req.UserID,
		Message:   req.Message,
		CreatedAt: req.Timestamp,
		Read:      false,
	}
	
	if err := h.db.WithContext(ctx).Create(&notification).Error; err != nil {
		log.Printf("Failed to save notification: %v", err)
	}
}

// Data structures for audit logging and notifications

type TeamAuditLog struct {
	ID          uint                   `gorm:"primaryKey"`
	TeamID      uuid.UUID              `gorm:"not null;index"`
	EventType   string                 `gorm:"not null"`
	PerformedBy uuid.UUID              `gorm:"not null"`
	Details     map[string]interface{} `gorm:"type:jsonb"`
	Timestamp   time.Time              `gorm:"not null"`
	CreatedAt   time.Time              `gorm:"autoCreateTime"`
}

type NotificationRequest struct {
	Type      string
	TeamID    uuid.UUID
	UserID    uuid.UUID
	Message   string
	Timestamp time.Time
}

type Notification struct {
	ID        uint      `gorm:"primaryKey"`
	Type      string    `gorm:"not null"`
	TeamID    uuid.UUID `gorm:"index"`
	UserID    uuid.UUID `gorm:"not null;index"`
	Message   string    `gorm:"not null"`
	Read      bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// Add these imports at the top
import (
	"fmt"
	"github.com/google/uuid"
)