package handler

import (
	"asset-management-api/internal/middleware"
	"asset-management-api/internal/service/interfaces"
	"asset-management-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamHandler struct {
	teamService interfaces.TeamService
}

type CreateTeamRequest struct {
	TeamName string `json:"teamName" validate:"required,min=1,max=255"`
	Managers []interfaces.TeamMemberInfo `json:"managers" validate:"dive"`
	Members  []interfaces.TeamMemberInfo `json:"members" validate:"dive"`
}

type AddMemberRequest struct {
	UserID   string `json:"userId" validate:"required,uuid"`
	UserName string `json:"userName" validate:"required"`
}

func NewTeamHandler(teamService interfaces.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// POST /teams
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user is manager
	userRole, _ := middleware.GetUserRoleFromContext(c)
	if userRole != "manager" {
		utils.ForbiddenResponse(c, "Only managers can create teams")
		return
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, utils.GetValidationErrorMessages(errors))
		return
	}

	team, err := h.teamService.CreateTeam(userID, req.TeamName, req.Managers, req.Members)
	if err != nil {
		if err.Error() == "access denied: only managers can create teams" {
			utils.ForbiddenResponse(c, "Manager role required")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to create team", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Team created successfully", team)
}

// POST /teams/:teamId/members
func (h *TeamHandler) AddMember(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID format", err)
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, utils.GetValidationErrorMessages(errors))
		return
	}

	memberID, err := uuid.Parse(req.UserID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.teamService.AddMember(teamID, userID, memberID)
	if err != nil {
		if err.Error() == "access denied: only team managers can add members" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		if err.Error() == "user is already a member of this team" {
			utils.BadRequestResponse(c, "User is already a team member", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to add member", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Member added successfully", nil)
}

// DELETE /teams/:teamId/members/:memberId
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID format", err)
		return
	}

	memberIDStr := c.Param("memberId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid member ID format", err)
		return
	}

	err = h.teamService.RemoveMember(teamID, userID, memberID)
	if err != nil {
		if err.Error() == "access denied: only team managers can remove members" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		if err.Error() == "member not found in team" {
			utils.NotFoundResponse(c, "Member not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to remove member", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Member removed successfully", nil)
}

// POST /teams/:teamId/managers
func (h *TeamHandler) AddManager(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID format", err)
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, utils.GetValidationErrorMessages(errors))
		return
	}

	managerID, err := uuid.Parse(req.UserID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID format", err)
		return
	}

	err = h.teamService.AddManager(teamID, userID, managerID)
	if err != nil {
		if err.Error() == "access denied: only team managers can add other managers" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		if err.Error() == "user is already a manager of this team" {
			utils.BadRequestResponse(c, "User is already a team manager", err)
			return
		}
		if err.Error() == "target user must have manager role" {
			utils.BadRequestResponse(c, "User must have manager role", err)
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to add manager", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Manager added successfully", nil)
}

// DELETE /teams/:teamId/managers/:managerId
func (h *TeamHandler) RemoveManager(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID format", err)
		return
	}

	managerIDStr := c.Param("managerId")
	managerID, err := uuid.Parse(managerIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid manager ID format", err)
		return
	}

	err = h.teamService.RemoveManager(teamID, userID, managerID)
	if err != nil {
		if err.Error() == "access denied: only team managers can remove other managers" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		if err.Error() == "manager not found in team" {
			utils.NotFoundResponse(c, "Manager not found")
			return
		}
		if err.Error() == "cannot remove the team creator" {
			utils.ForbiddenResponse(c, "Cannot remove team creator")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to remove manager", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Manager removed successfully", nil)
}

// GET /teams/:teamId
func (h *TeamHandler) GetTeam(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	teamIDStr := c.Param("teamId")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID format", err)
		return
	}

	team, err := h.teamService.GetTeam(teamID, userID)
	if err != nil {
		if err.Error() == "access denied: you are not a member of this team" {
			utils.ForbiddenResponse(c, "Access denied")
			return
		}
		if err.Error() == "team not found" {
			utils.NotFoundResponse(c, "Team not found")
			return
		}
		utils.InternalServerErrorResponse(c, "Failed to get team", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Team retrieved successfully", team)
}

// GET /teams (Get user's teams)
func (h *TeamHandler) GetUserTeams(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	teams, err := h.teamService.GetUserTeams(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get teams", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Teams retrieved successfully", teams)
}